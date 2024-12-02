package writecache

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/epicchainlabs/neofs-node/pkg/local_object_storage/blobstor"
	"github.com/epicchainlabs/neofs-node/pkg/local_object_storage/blobstor/fstree"
	"github.com/epicchainlabs/neofs-node/pkg/local_object_storage/internal/storagetest"
	meta "github.com/epicchainlabs/neofs-node/pkg/local_object_storage/metabase"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGeneric(t *testing.T) {
	defer func() { _ = os.RemoveAll(t.Name()) }()

	var n int
	newCache := func(t *testing.T) storagetest.Component {
		n++
		dir := filepath.Join(t.Name(), strconv.Itoa(n))
		require.NoError(t, os.MkdirAll(dir, os.ModePerm))
		return New(
			WithLogger(zaptest.NewLogger(t)),
			WithFlushWorkersCount(2),
			WithPath(dir))
	}

	storagetest.TestAll(t, newCache)
}

func newCache(tb testing.TB, smallSize uint64, opts ...Option) (Cache, *blobstor.BlobStor, *meta.DB) {
	dir := tb.TempDir()
	mb := meta.New(
		meta.WithPath(filepath.Join(dir, "meta")),
		meta.WithEpochState(dummyEpoch{}))
	require.NoError(tb, mb.Open(false))
	require.NoError(tb, mb.Init())

	fsTree := fstree.New(
		fstree.WithPath(filepath.Join(dir, "blob")),
		fstree.WithDepth(0),
		fstree.WithDirNameLen(1))
	bs := blobstor.New(
		blobstor.WithStorages([]blobstor.SubStorage{{Storage: fsTree}}),
		blobstor.WithCompressObjects(true))
	require.NoError(tb, bs.Open(false))
	require.NoError(tb, bs.Init())

	wc := New(
		append([]Option{
			WithPath(filepath.Join(dir, "writecache")),
			WithSmallObjectSize(smallSize),
			WithMetabase(mb),
			WithBlobstor(bs),
		}, opts...)...)
	require.NoError(tb, wc.Open(false))
	require.NoError(tb, wc.Init())

	return wc, bs, mb
}
