package policer

import (
	"sync"
	"time"

	"github.com/epicchainlabs/epicchain-node/pkg/core/container"
	"github.com/epicchainlabs/epicchain-node/pkg/core/netmap"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/engine"
	headsvc "github.com/epicchainlabs/epicchain-node/pkg/services/object/head"
	"github.com/epicchainlabs/epicchain-node/pkg/services/object_manager/placement"
	"github.com/epicchainlabs/epicchain-node/pkg/services/replicator"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

// NodeLoader provides application load statistics.
type nodeLoader interface {
	// ObjectServiceLoad returns object service load value in [0:1] range.
	ObjectServiceLoad() float64
}

type objectsInWork struct {
	m    sync.RWMutex
	objs map[oid.Address]struct{}
}

func (oiw *objectsInWork) inWork(addr oid.Address) bool {
	oiw.m.RLock()
	_, ok := oiw.objs[addr]
	oiw.m.RUnlock()

	return ok
}

func (oiw *objectsInWork) remove(addr oid.Address) {
	oiw.m.Lock()
	delete(oiw.objs, addr)
	oiw.m.Unlock()
}

func (oiw *objectsInWork) add(addr oid.Address) {
	oiw.m.Lock()
	oiw.objs[addr] = struct{}{}
	oiw.m.Unlock()
}

// Policer represents the utility that verifies
// compliance with the object storage policy.
type Policer struct {
	*cfg

	objsInWork *objectsInWork
}

// Option is an option for Policer constructor.
type Option func(*cfg)

// RedundantCopyCallback is a callback to pass
// the redundant local copy of the object.
type RedundantCopyCallback func(oid.Address)

// Network provides information about the NeoFS network to Policer for work.
type Network interface {
	// IsLocalNodeInNetmap checks whether the local node belongs to the current
	// network map. If it is impossible to check this fact, IsLocalNodeInNetmap
	// returns false.
	IsLocalNodeInNetmap() bool
}

type cfg struct {
	sync.RWMutex
	// available for runtime reconfiguration
	headTimeout time.Duration
	repCooldown time.Duration
	batchSize   uint32
	maxCapacity uint32

	log *zap.Logger

	jobQueue jobQueue

	cnrSrc container.Source

	placementBuilder placement.Builder

	remoteHeader *headsvc.RemoteHeader

	netmapKeys netmap.AnnouncedKeys

	replicator *replicator.Replicator

	cbRedundantCopy RedundantCopyCallback

	taskPool *ants.Pool

	loader nodeLoader

	rebalanceFreq time.Duration

	network Network
}

func defaultCfg() *cfg {
	return &cfg{
		log:           zap.L(),
		batchSize:     10,
		rebalanceFreq: 1 * time.Second,
		repCooldown:   1 * time.Second,
	}
}

// New creates, initializes and returns Policer instance.
func New(opts ...Option) *Policer {
	c := defaultCfg()

	for i := range opts {
		opts[i](c)
	}

	c.log = c.log.With(zap.String("component", "Object Policer"))

	return &Policer{
		cfg: c,
		objsInWork: &objectsInWork{
			objs: make(map[oid.Address]struct{}, c.maxCapacity),
		},
	}
}

// WithHeadTimeout returns option to set Head timeout of Policer.
func WithHeadTimeout(v time.Duration) Option {
	return func(c *cfg) {
		c.headTimeout = v
	}
}

// WithLogger returns option to set Logger of Policer.
func WithLogger(v *zap.Logger) Option {
	return func(c *cfg) {
		c.log = v
	}
}

// WithLocalStorage returns option to set local object storage of Policer.
func WithLocalStorage(v *engine.StorageEngine) Option {
	return func(c *cfg) {
		c.jobQueue.localStorage = v
	}
}

// WithContainerSource returns option to set container source of Policer.
func WithContainerSource(v container.Source) Option {
	return func(c *cfg) {
		c.cnrSrc = v
	}
}

// WithPlacementBuilder returns option to set object placement builder of Policer.
func WithPlacementBuilder(v placement.Builder) Option {
	return func(c *cfg) {
		c.placementBuilder = v
	}
}

// WithRemoteHeader returns option to set object header receiver of Policer.
func WithRemoteHeader(v *headsvc.RemoteHeader) Option {
	return func(c *cfg) {
		c.remoteHeader = v
	}
}

// WithNetmapKeys returns option to set tool to work with announced public keys.
func WithNetmapKeys(v netmap.AnnouncedKeys) Option {
	return func(c *cfg) {
		c.netmapKeys = v
	}
}

// WithReplicator returns option to set object replicator of Policer.
func WithReplicator(v *replicator.Replicator) Option {
	return func(c *cfg) {
		c.replicator = v
	}
}

// WithRedundantCopyCallback returns option to set
// callback to pass redundant local object copies
// detected by Policer.
func WithRedundantCopyCallback(cb RedundantCopyCallback) Option {
	return func(c *cfg) {
		c.cbRedundantCopy = cb
	}
}

// WithMaxCapacity returns option to set max capacity
// that can be set to the pool.
func WithMaxCapacity(capacity uint32) Option {
	return func(c *cfg) {
		c.maxCapacity = capacity
	}
}

// WithPool returns option to set pool for
// policy and replication operations.
func WithPool(p *ants.Pool) Option {
	return func(c *cfg) {
		c.taskPool = p
	}
}

// WithNodeLoader returns option to set NeoFS node load source.
func WithNodeLoader(l nodeLoader) Option {
	return func(c *cfg) {
		c.loader = l
	}
}

// WithNetwork provides Network component.
func WithNetwork(n Network) Option {
	return func(c *cfg) {
		c.network = n
	}
}

// WithReplicationCooldown returns option to set replication
// cooldown: the [Policer] will not submit more than one task
// per a provided time duration.
func WithReplicationCooldown(d time.Duration) Option {
	return func(c *cfg) {
		c.repCooldown = d
	}
}

// WithObjectBatchSize returns option to set maximum objects
// read from the Storage at once.
func WithObjectBatchSize(s uint32) Option {
	return func(c *cfg) {
		c.batchSize = s
	}
}
