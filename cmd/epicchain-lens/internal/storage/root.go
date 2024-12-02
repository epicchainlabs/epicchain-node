package storage

import (
	"fmt"
	"time"

	common "github.com/epicchainlabs/epicchain-node/cmd/epicchain-lens/internal"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config"
	engineconfig "github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config/engine"
	shardconfig "github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config/engine/shard"
	fstreeconfig "github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config/engine/shard/blobstor/fstree"
	peapodconfig "github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config/engine/shard/blobstor/peapod"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/storage"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/blobstor"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/blobstor/fstree"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/blobstor/peapod"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/engine"
	meta "github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/metabase"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/pilorama"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/shard"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/shard/mode"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/writecache"
	"github.com/epicchainlabs/epicchain-node/pkg/util"
	objectSDK "github.com/epicchainlabs/epicchain-sdk-go/object"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
)

var (
	vAddress     string
	vOut         string
	vConfig      string
	vPayloadOnly bool
)

var Root = &cobra.Command{
	Use:   "storage",
	Short: "Operations with blobstor",
	Args:  cobra.NoArgs,
}

func init() {
	Root.AddCommand(
		storageGetObjCMD,
		storageListObjsCMD,
		storageStatusObjCMD,
	)
}

type shardOptsWithID struct {
	configID string
	shOpts   []shard.Option
}

type epochState struct {
}

func (e epochState) CurrentEpoch() uint64 {
	return 0
}

func openEngine(cmd *cobra.Command) *engine.StorageEngine {
	appCfg := config.New(config.Prm{}, config.WithConfigFile(vConfig))

	ls := engine.New()

	var shards []storage.ShardCfg
	err := engineconfig.IterateShards(appCfg, false, func(sc *shardconfig.Config) error {
		var sh storage.ShardCfg

		sh.RefillMetabase = sc.RefillMetabase()
		sh.Mode = sc.Mode()
		sh.Compress = sc.Compress()
		sh.UncompressableContentType = sc.UncompressableContentTypes()
		sh.SmallSizeObjectLimit = sc.SmallSizeLimit()

		// write-cache

		writeCacheCfg := sc.WriteCache()
		if writeCacheCfg.Enabled() {
			wc := &sh.WritecacheCfg

			wc.Enabled = true
			wc.Path = writeCacheCfg.Path()
			wc.MaxBatchSize = writeCacheCfg.BoltDB().MaxBatchSize()
			wc.MaxBatchDelay = writeCacheCfg.BoltDB().MaxBatchDelay()
			wc.MaxObjSize = writeCacheCfg.MaxObjectSize()
			wc.SmallObjectSize = writeCacheCfg.SmallObjectSize()
			wc.FlushWorkerCount = writeCacheCfg.WorkersNumber()
			wc.SizeLimit = writeCacheCfg.SizeLimit()
			wc.NoSync = writeCacheCfg.NoSync()
		}

		// blobstor with substorages

		blobStorCfg := sc.BlobStor()
		storagesCfg := blobStorCfg.Storages()
		metabaseCfg := sc.Metabase()
		gcCfg := sc.GC()

		if config.BoolSafe(appCfg.Sub("tree"), "enabled") {
			piloramaCfg := sc.Pilorama()
			pr := &sh.PiloramaCfg

			pr.Enabled = true
			pr.Path = piloramaCfg.Path()
			pr.Perm = piloramaCfg.Perm()
			pr.NoSync = piloramaCfg.NoSync()
			pr.MaxBatchSize = piloramaCfg.MaxBatchSize()
			pr.MaxBatchDelay = piloramaCfg.MaxBatchDelay()
		}

		ss := make([]storage.SubStorageCfg, 0, len(storagesCfg))
		for i := range storagesCfg {
			var sCfg storage.SubStorageCfg

			sCfg.Typ = storagesCfg[i].Type()
			sCfg.Path = storagesCfg[i].Path()
			sCfg.Perm = storagesCfg[i].Perm()

			switch storagesCfg[i].Type() {
			case fstree.Type:
				sub := fstreeconfig.From((*config.Config)(storagesCfg[i]))
				sCfg.Depth = sub.Depth()
				sCfg.NoSync = sub.NoSync()
			case peapod.Type:
				peapodCfg := peapodconfig.From((*config.Config)(storagesCfg[i]))
				sCfg.FlushInterval = peapodCfg.FlushInterval()
			default:
				return fmt.Errorf("can't initiate storage. invalid storage type: %s", storagesCfg[i].Type())
			}

			ss = append(ss, sCfg)
		}

		sh.SubStorages = ss

		// meta

		m := &sh.MetaCfg

		m.Path = metabaseCfg.Path()
		m.Perm = metabaseCfg.BoltDB().Perm()
		m.MaxBatchDelay = metabaseCfg.BoltDB().MaxBatchDelay()
		m.MaxBatchSize = metabaseCfg.BoltDB().MaxBatchSize()

		// GC

		sh.GcCfg.RemoverBatchSize = gcCfg.RemoverBatchSize()
		sh.GcCfg.RemoverSleepInterval = gcCfg.RemoverSleepInterval()

		shards = append(shards, sh)

		return nil
	})
	common.ExitOnErr(cmd, err)

	var shardsWithMeta []shardOptsWithID
	for _, shCfg := range shards {
		var writeCacheOpts []writecache.Option
		if wcRead := shCfg.WritecacheCfg; wcRead.Enabled {
			writeCacheOpts = append(writeCacheOpts,
				writecache.WithPath(wcRead.Path),
				writecache.WithMaxBatchSize(wcRead.MaxBatchSize),
				writecache.WithMaxBatchDelay(wcRead.MaxBatchDelay),
				writecache.WithMaxObjectSize(wcRead.MaxObjSize),
				writecache.WithSmallObjectSize(wcRead.SmallObjectSize),
				writecache.WithFlushWorkersCount(wcRead.FlushWorkerCount),
				writecache.WithMaxCacheSize(wcRead.SizeLimit),
				writecache.WithNoSync(wcRead.NoSync),
			)
		}

		var piloramaOpts []pilorama.Option
		if prRead := shCfg.PiloramaCfg; prRead.Enabled {
			piloramaOpts = append(piloramaOpts,
				pilorama.WithPath(prRead.Path),
				pilorama.WithPerm(prRead.Perm),
				pilorama.WithNoSync(prRead.NoSync),
				pilorama.WithMaxBatchSize(prRead.MaxBatchSize),
				pilorama.WithMaxBatchDelay(prRead.MaxBatchDelay),
			)
		}

		var ss []blobstor.SubStorage
		for _, sRead := range shCfg.SubStorages {
			switch sRead.Typ {
			case fstree.Type:
				ss = append(ss, blobstor.SubStorage{
					Storage: fstree.New(
						fstree.WithPath(sRead.Path),
						fstree.WithPerm(sRead.Perm),
						fstree.WithDepth(sRead.Depth),
						fstree.WithNoSync(sRead.NoSync)),
					Policy: func(_ *objectSDK.Object, data []byte) bool {
						return true
					},
				})
			case peapod.Type:
				ss = append(ss, blobstor.SubStorage{
					Storage: peapod.New(sRead.Path, sRead.Perm, sRead.FlushInterval),
					Policy: func(_ *objectSDK.Object, data []byte) bool {
						return uint64(len(data)) < shCfg.SmallSizeObjectLimit
					},
				})
			default:
				// should never happen, that has already
				// been handled: when the config was read
			}
		}

		var sh shardOptsWithID
		sh.configID = shCfg.ID()
		sh.shOpts = []shard.Option{
			shard.WithRefillMetabase(shCfg.RefillMetabase),
			shard.WithMode(shCfg.Mode),
			shard.WithBlobStorOptions(
				blobstor.WithCompressObjects(shCfg.Compress),
				blobstor.WithUncompressableContentTypes(shCfg.UncompressableContentType),
				blobstor.WithStorages(ss),
			),
			shard.WithMetaBaseOptions(
				meta.WithPath(shCfg.MetaCfg.Path),
				meta.WithPermissions(shCfg.MetaCfg.Perm),
				meta.WithMaxBatchSize(shCfg.MetaCfg.MaxBatchSize),
				meta.WithMaxBatchDelay(shCfg.MetaCfg.MaxBatchDelay),
				meta.WithBoltDBOptions(&bbolt.Options{
					Timeout: time.Second,
				}),

				meta.WithEpochState(epochState{}),
			),
			shard.WithPiloramaOptions(piloramaOpts...),
			shard.WithWriteCache(shCfg.WritecacheCfg.Enabled),
			shard.WithWriteCacheOptions(writeCacheOpts...),
			shard.WithRemoverBatchSize(shCfg.GcCfg.RemoverBatchSize),
			shard.WithGCRemoverSleepInterval(shCfg.GcCfg.RemoverSleepInterval),
			shard.WithGCWorkerPoolInitializer(func(sz int) util.WorkerPool {
				pool, err := ants.NewPool(sz)
				common.ExitOnErr(cmd, err)

				return pool
			}),
		}

		shardsWithMeta = append(shardsWithMeta, sh)
	}

	for _, optsWithMeta := range shardsWithMeta {
		_, err := ls.AddShard(append(optsWithMeta.shOpts, shard.WithMode(mode.ReadOnly))...)
		common.ExitOnErr(cmd, err)
	}

	common.ExitOnErr(cmd, ls.Open())
	common.ExitOnErr(cmd, ls.Init())

	return ls
}
