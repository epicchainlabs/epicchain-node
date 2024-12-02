package engine

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/epicchainlabs/neofs-node/pkg/core/container"
	"github.com/epicchainlabs/neofs-node/pkg/core/object"
	"github.com/epicchainlabs/neofs-node/pkg/local_object_storage/blobstor"
	meta "github.com/epicchainlabs/neofs-node/pkg/local_object_storage/metabase"
	"github.com/epicchainlabs/neofs-node/pkg/local_object_storage/shard"
	apistatus "github.com/epicchainlabs/neofs-sdk-go/client/status"
	cid "github.com/epicchainlabs/neofs-sdk-go/container/id"
	objecttest "github.com/epicchainlabs/neofs-sdk-go/object/test"
	"github.com/stretchr/testify/require"
)

type cnrSource struct{}

func (c cnrSource) Get(cid.ID) (*container.Container, error) {
	return nil, apistatus.ContainerNotFound{} // value not used, only err
}

func TestStorageEngine_ContainerCleanUp(t *testing.T) {
	path := t.TempDir()

	e := New(WithContainersSource(cnrSource{}))
	t.Cleanup(func() {
		_ = e.Close()
	})

	for i := 0; i < 5; i++ {
		_, err := e.AddShard(
			shard.WithBlobStorOptions(
				blobstor.WithStorages(newStorages(filepath.Join(path, strconv.Itoa(i)), errSmallSize))),
			shard.WithMetaBaseOptions(
				meta.WithPath(filepath.Join(path, fmt.Sprintf("%d.metabase", i))),
				meta.WithPermissions(0700),
				meta.WithEpochState(epochState{}),
			),
		)
		require.NoError(t, err)
	}
	require.NoError(t, e.Open())

	o1 := objecttest.Object(t)
	o2 := objecttest.Object(t)
	o2.SetPayload(make([]byte, errSmallSize+1))

	var prmPut PutPrm
	prmPut.WithObject(&o1)

	_, err := e.Put(prmPut)
	require.NoError(t, err)

	prmPut.WithObject(&o2)
	_, err = e.Put(prmPut)
	require.NoError(t, err)

	require.NoError(t, e.Init())

	require.Eventually(t, func() bool {
		var prmGet GetPrm
		prmGet.WithAddress(object.AddressOf(&o1))
		_, err1 := e.Get(prmGet)

		prmGet.WithAddress(object.AddressOf(&o2))
		_, err2 := e.Get(prmGet)

		return errors.Is(err1, new(apistatus.ObjectNotFound)) && errors.Is(err2, new(apistatus.ObjectNotFound))
	}, time.Second, 100*time.Millisecond)
}
