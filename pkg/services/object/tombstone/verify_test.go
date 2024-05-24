package tombstone

import (
	"context"
	"testing"

	objectcore "github.com/nspcc-dev/neofs-node/pkg/core/object"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/stretchr/testify/require"
)

type headRes struct {
	h   *object.Object
	err error
}

type testObjectSource struct {
	searchV1 map[object.SplitID][]oid.ID
	searchV2 map[oid.ID][]oid.ID
	head     map[oid.Address]headRes
}

func (t *testObjectSource) Head(_ context.Context, addr oid.Address) (*object.Object, error) {
	res := t.head[addr]
	return res.h, res.err
}

func (t *testObjectSource) Search(_ context.Context, _ cid.ID, ff object.SearchFilters) ([]oid.ID, error) {
	f := ff[0]

	switch f.Header() {
	case object.FilterSplitID:
		if t.searchV1 == nil {
			return nil, nil
		}

		var splitID object.SplitID
		err := splitID.Parse(f.Value())
		if err != nil {
			panic(err)
		}

		return t.searchV1[splitID], nil
	case object.FilterFirstSplitObject:
		if t.searchV2 == nil {
			return nil, nil
		}

		var firstObject oid.ID
		err := firstObject.DecodeString(f.Value())
		if err != nil {
			panic(err)
		}

		return t.searchV2[firstObject], nil
	default:
		panic("unexpected search call")
	}
}

func TestVerifier_VerifyTomb(t *testing.T) {
	os := &testObjectSource{}
	ctx := context.Background()

	v := NewVerifier(os)

	t.Run("with splitID", func(t *testing.T) {
		var tomb object.Tombstone
		tomb.SetSplitID(object.NewSplitID())

		err := v.VerifyTomb(ctx, cid.ID{}, tomb)
		require.ErrorContains(t, err, "unexpected split")
	})

	t.Run("tombs with small objects", func(t *testing.T) {
		var tomb object.Tombstone

		cnr := cidtest.ID()
		children := []object.Object{
			objectWithCnr(t, cnr, false),
			objectWithCnr(t, cnr, false),
			objectWithCnr(t, cnr, false),
		}

		*os = testObjectSource{
			head: childrenResMap(cnr, children),
		}

		tomb.SetMembers(objectsToOIDs(children))

		require.NoError(t, v.VerifyTomb(ctx, cnr, tomb))
	})

	t.Run("tomb with children", func(t *testing.T) {
		var tomb object.Tombstone

		cnr := cidtest.ID()
		child := objectWithCnr(t, cnr, true)
		childID, _ := child.ID()
		splitID := child.SplitID()

		var addr oid.Address
		addr.SetContainer(cnr)
		addr.SetObject(childID)

		tomb.SetMembers([]oid.ID{childID})

		t.Run("V1", func(t *testing.T) {
			t.Run("LINKs can not be found", func(t *testing.T) {
				*os = testObjectSource{
					head: map[oid.Address]headRes{
						addr: {
							h: &child,
						},
					},
				}

				require.NoError(t, v.VerifyTomb(ctx, cnr, tomb))
			})

			t.Run("LINKs can be found", func(t *testing.T) {
				link := objectWithCnr(t, cnr, false)
				link.SetChildren(childID)
				linkID, _ := link.ID()

				objectcore.AddressOf(&link)

				*os = testObjectSource{
					head: map[oid.Address]headRes{
						addr: {
							h: &child,
						},
						objectcore.AddressOf(&link): {
							h: &link,
						},
					},
					searchV1: map[object.SplitID][]oid.ID{
						*splitID: {linkID},
					},
				}

				err := v.VerifyTomb(ctx, cnr, tomb)
				require.ErrorContains(t, err, "V1")
				require.ErrorContains(t, err, "found link object")
			})
		})

		t.Run("V2", func(t *testing.T) {
			child.SetSplitID(nil)

			t.Run("removing first object", func(t *testing.T) {
				*os = testObjectSource{
					head: map[oid.Address]headRes{
						addr: {
							h: &child,
						},
					},
					searchV2: map[oid.ID][]oid.ID{
						childID: {oidtest.ID()}, // the first object is a chain ID in itself
					},
				}

				err := v.VerifyTomb(ctx, cnr, tomb)
				require.ErrorContains(t, err, "V2")
				require.ErrorContains(t, err, "found link object")
			})

			firstObject := oidtest.ID()
			child.SetFirstID(firstObject)

			t.Run("LINKs can not be found", func(t *testing.T) {
				*os = testObjectSource{
					head: map[oid.Address]headRes{
						addr: {
							h: &child,
						},
					},
				}

				require.NoError(t, v.VerifyTomb(ctx, cnr, tomb))
			})

			t.Run("LINKs can be found", func(t *testing.T) {
				*os = testObjectSource{
					head: map[oid.Address]headRes{
						addr: {
							h: &child,
						},
					},
					searchV2: map[oid.ID][]oid.ID{
						firstObject: {oidtest.ID()},
					},
				}

				err := v.VerifyTomb(ctx, cnr, tomb)
				require.ErrorContains(t, err, "V2")
				require.ErrorContains(t, err, "found link object")
			})
		})
	})

	t.Run("tomb with parent", func(t *testing.T) {
		addr := oidtest.Address()
		si := object.NewSplitInfo()
		siErr := object.NewSplitInfoError(si)

		var tomb object.Tombstone
		tomb.SetMembers([]oid.ID{addr.Object()})

		*os = testObjectSource{
			head: map[oid.Address]headRes{
				addr: {
					err: siErr,
				},
			},
		}

		require.NoError(t, v.VerifyTomb(ctx, cidtest.ID(), tomb))
	})
}

func childrenResMap(cnr cid.ID, heads []object.Object) map[oid.Address]headRes {
	res := make(map[oid.Address]headRes)

	var addr oid.Address
	addr.SetContainer(cnr)

	for _, obj := range heads {
		oID, _ := obj.ID()
		addr.SetObject(oID)

		obj.SetContainerID(cnr)

		res[addr] = headRes{
			h:   &obj,
			err: nil,
		}
	}

	return res
}

func objectsToOIDs(oo []object.Object) []oid.ID {
	res := make([]oid.ID, len(oo))
	for _, obj := range oo {
		oID, _ := obj.ID()
		res = append(res, oID)
	}

	return res
}

func objectWithCnr(t *testing.T, cnr cid.ID, hasParent bool) object.Object {
	obj := objecttest.Object(t)
	obj.SetContainerID(cnr)

	if !hasParent {
		obj.ResetRelations()
	}

	return obj
}
