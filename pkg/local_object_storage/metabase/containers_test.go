package meta_test

import (
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/epicchainlabs/epicchain-node/pkg/core/object"
	cid "github.com/epicchainlabs/epicchain-sdk-go/container/id"
	cidtest "github.com/epicchainlabs/epicchain-sdk-go/container/id/test"
	objectSDK "github.com/epicchainlabs/epicchain-sdk-go/object"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
	oidtest "github.com/epicchainlabs/epicchain-sdk-go/object/id/test"
	objecttest "github.com/epicchainlabs/epicchain-sdk-go/object/test"
	"github.com/stretchr/testify/require"
)

func TestDB_Containers(t *testing.T) {
	db := newDB(t)

	const N = 10

	cids := make(map[string]int, N)

	for i := 0; i < N; i++ {
		obj := generateObject(t)

		cnr, _ := obj.ContainerID()

		cids[cnr.EncodeToString()] = 0

		err := putBig(db, obj)
		require.NoError(t, err)
	}

	lst, err := db.Containers()
	require.NoError(t, err)

	for _, cnr := range lst {
		i, ok := cids[cnr.EncodeToString()]
		require.True(t, ok)
		require.Equal(t, 0, i)

		cids[cnr.EncodeToString()] = 1
	}

	// require.Contains not working since cnrs is a ptr slice
	assertContains := func(cnrs []cid.ID, cnr cid.ID) {
		found := false
		for i := 0; !found && i < len(cnrs); i++ {
			found = cnrs[i].Equals(cnr)
		}

		require.True(t, found)
	}

	t.Run("Inhume", func(t *testing.T) {
		obj := generateObject(t)

		require.NoError(t, putBig(db, obj))

		cnrs, err := db.Containers()
		require.NoError(t, err)
		cnr, _ := obj.ContainerID()

		assertContains(cnrs, cnr)

		require.NoError(t, metaInhume(db, object.AddressOf(obj), oidtest.Address()))

		cnrs, err = db.Containers()
		require.NoError(t, err)
		assertContains(cnrs, cnr)
	})

	t.Run("ToMoveIt", func(t *testing.T) {
		obj := generateObject(t)

		require.NoError(t, putBig(db, obj))

		cnrs, err := db.Containers()
		require.NoError(t, err)
		cnr, _ := obj.ContainerID()
		assertContains(cnrs, cnr)

		require.NoError(t, metaToMoveIt(db, object.AddressOf(obj)))

		cnrs, err = db.Containers()
		require.NoError(t, err)
		assertContains(cnrs, cnr)
	})
}

func TestDB_ContainersCount(t *testing.T) {
	db := newDB(t)

	const R, T, SG, L = 10, 11, 12, 13 // amount of object per type

	uploadObjects := [...]struct {
		amount int
		typ    objectSDK.Type
	}{
		{R, objectSDK.TypeRegular},
		{T, objectSDK.TypeTombstone},
		{SG, objectSDK.TypeStorageGroup},
		{L, objectSDK.TypeLock},
	}

	expected := make([]cid.ID, 0, R+T+SG+L)

	for _, upload := range uploadObjects {
		for i := 0; i < upload.amount; i++ {
			obj := generateObject(t)
			obj.SetType(upload.typ)

			err := putBig(db, obj)
			require.NoError(t, err)

			cnr, _ := obj.ContainerID()
			expected = append(expected, cnr)
		}
	}

	sort.Slice(expected, func(i, j int) bool {
		return expected[i].EncodeToString() < expected[j].EncodeToString()
	})

	got, err := db.Containers()
	require.NoError(t, err)

	sort.Slice(got, func(i, j int) bool {
		return got[i].EncodeToString() < got[j].EncodeToString()
	})

	require.Equal(t, expected, got)
}

func TestDB_ContainerSize(t *testing.T) {
	db := newDB(t)

	const (
		C = 3
		N = 5
	)

	cids := make(map[cid.ID]int, C)
	objs := make(map[cid.ID][]*objectSDK.Object, C*N)

	for i := 0; i < C; i++ {
		cnr := cidtest.ID()
		cids[cnr] = 0

		for j := 0; j < N; j++ {
			size := rand.Intn(1024)

			parent := generateObjectWithCID(t, cnr)
			parent.SetPayloadSize(uint64(size / 2))

			obj := generateObjectWithCID(t, cnr)
			obj.SetPayloadSize(uint64(size))
			idParent, _ := parent.ID()
			obj.SetParentID(idParent)
			obj.SetParent(parent)

			cids[cnr] += size
			objs[cnr] = append(objs[cnr], obj)

			err := putBig(db, obj)
			require.NoError(t, err)
		}
	}

	for cnr, volume := range cids {
		n, err := db.ContainerSize(cnr)
		require.NoError(t, err)
		require.Equal(t, volume, int(n))
	}

	t.Run("Inhume", func(t *testing.T) {
		for cnr, list := range objs {
			volume := cids[cnr]

			for _, obj := range list {
				require.NoError(t, metaInhume(
					db,
					object.AddressOf(obj),
					oidtest.Address(),
				))

				volume -= int(obj.PayloadSize())

				n, err := db.ContainerSize(cnr)
				require.NoError(t, err)
				require.Equal(t, volume, int(n))
			}
		}
	})
}

func TestDB_DeleteContainer(t *testing.T) {
	db := newDB(t)
	cID := cidtest.ID()

	t.Run("does not exist", func(t *testing.T) {
		err := db.DeleteContainer(cidtest.ID())
		require.NoError(t, err)
	})

	t.Run("exist", func(t *testing.T) {
		var attr objectSDK.Attribute
		attr.SetKey("test")
		attr.SetValue("test")

		o1 := generateObjectWithCID(t, cID)
		o1.SetAttributes(attr)
		storageID := []byte{1, 2, 3, 4}

		// put one object with storageID and an  attribute
		err := metaPut(db, o1, storageID)
		require.NoError(t, err)

		fetchedStorageID, err := metaStorageID(db, object.AddressOf(o1))
		require.NoError(t, err)
		require.Equal(t, storageID, fetchedStorageID)

		// put a big one
		o2 := objecttest.Object(t)
		o2.SetContainerID(cID)
		err = putBig(db, &o2)
		require.NoError(t, err)

		// lockers
		err = db.Lock(cID, oidtest.ID(), []oid.ID{oidtest.ID()})
		require.NoError(t, err)

		// SG
		o3 := objecttest.Object(t)
		o3.SetContainerID(cID)
		o3.SetType(objectSDK.TypeStorageGroup)
		err = putBig(db, &o3)
		require.NoError(t, err)

		// TS
		o4 := objecttest.Object(t)
		o4.SetContainerID(cID)
		o4.SetType(objectSDK.TypeTombstone)
		err = metaInhume(db, object.AddressOf(o1), object.AddressOf(&o4))
		require.NoError(t, err)

		err = db.DeleteContainer(cID)
		require.NoError(t, err)

		objs, err := db.ListContainerObjects(cID, math.MaxInt64)
		require.NoError(t, err)

		require.Len(t, objs, 0)
	})
}
