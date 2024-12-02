package writecache_test

import (
	"testing"

	objectcore "github.com/epicchainlabs/epicchain-node/pkg/core/object"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/blobstor/common"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/writecache"
	objecttest "github.com/epicchainlabs/epicchain-sdk-go/object/test"
	"github.com/stretchr/testify/require"
)

func TestCache_InitReadOnly(t *testing.T) {
	dir := t.TempDir()

	wc := writecache.New(
		writecache.WithPath(dir),
	)

	// open in rw mode first to create underlying BoltDB with some object
	// (otherwise 'bad file descriptor' error on Open occurs)
	err := wc.Open(false)
	require.NoError(t, err)

	err = wc.Init()
	require.NoError(t, err)

	obj := objecttest.Object(t)

	_, err = wc.Put(common.PutPrm{
		Address: objectcore.AddressOf(&obj),
		Object:  &obj,
	})
	require.NoError(t, err)

	err = wc.Close()
	require.NoError(t, err)

	// try Init in read-only mode
	err = wc.Open(true)
	require.NoError(t, err)

	t.Cleanup(func() { wc.Close() })

	err = wc.Init()
	require.NoError(t, err)
}
