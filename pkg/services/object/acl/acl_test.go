package acl

import (
	"testing"

	"github.com/epicchainlabs/epicchain-node/pkg/core/container"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/engine"
	v2 "github.com/epicchainlabs/epicchain-node/pkg/services/object/acl/v2"
	"github.com/epicchainlabs/epicchain-sdk-go/container/acl"
	cid "github.com/epicchainlabs/epicchain-sdk-go/container/id"
	eaclSDK "github.com/epicchainlabs/epicchain-sdk-go/eacl"
	"github.com/epicchainlabs/epicchain-sdk-go/object"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
	"github.com/epicchainlabs/epicchain-sdk-go/user"
	usertest "github.com/epicchainlabs/epicchain-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

type emptyEACLSource struct{}

func (e emptyEACLSource) GetEACL(_ cid.ID) (*container.EACL, error) {
	return nil, nil
}

type emptyNetmapState struct{}

type emptyHeaderSource struct{}

func (e emptyHeaderSource) Head(address oid.Address) (*object.Object, error) {
	return nil, nil
}

func (e emptyNetmapState) CurrentEpoch() uint64 {
	return 0
}

func TestStickyCheck(t *testing.T) {
	checker := NewChecker(new(CheckerPrm).
		SetLocalStorage(&engine.StorageEngine{}).
		SetValidator(eaclSDK.NewValidator()).
		SetEACLSource(emptyEACLSource{}).
		SetNetmapState(emptyNetmapState{}).
		SetHeaderSource(emptyHeaderSource{}),
	)

	t.Run("system role", func(t *testing.T) {
		var info v2.RequestInfo

		info.SetSenderKey(make([]byte, 33)) // any non-empty key
		info.SetRequestRole(acl.RoleContainer)

		require.True(t, checker.StickyBitCheck(info, usertest.ID(t)))

		var basicACL acl.Basic
		basicACL.MakeSticky()

		info.SetBasicACL(basicACL)

		require.True(t, checker.StickyBitCheck(info, usertest.ID(t)))
	})

	t.Run("owner ID and/or public key emptiness", func(t *testing.T) {
		var info v2.RequestInfo

		info.SetRequestRole(acl.RoleOthers) // should be non-system role

		assertFn := func(isSticky, withKey, withOwner, expected bool) {
			info := info
			if isSticky {
				var basicACL acl.Basic
				basicACL.MakeSticky()

				info.SetBasicACL(basicACL)
			}

			if withKey {
				info.SetSenderKey(make([]byte, 33))
			} else {
				info.SetSenderKey(nil)
			}

			var ownerID user.ID

			if withOwner {
				ownerID = usertest.ID(t)
			}

			require.Equal(t, expected, checker.StickyBitCheck(info, ownerID))
		}

		assertFn(true, false, false, false)
		assertFn(true, true, false, false)
		assertFn(true, false, true, false)
		assertFn(false, false, false, true)
		assertFn(false, true, false, true)
		assertFn(false, false, true, true)
		assertFn(false, true, true, true)
	})
}
