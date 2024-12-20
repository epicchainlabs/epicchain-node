package control

import (
	"crypto/ecdsa"
	"fmt"
	"sync/atomic"

	"github.com/epicchainlabs/epicchain-node/pkg/core/container"
	"github.com/epicchainlabs/epicchain-node/pkg/core/netmap"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/engine"
	"github.com/epicchainlabs/epicchain-node/pkg/services/control"
	"github.com/epicchainlabs/epicchain-node/pkg/services/replicator"
)

// Server is an entity that serves
// Control service on storage node.
type Server struct {
	// initialization sync; locks any calls except
	// health checks before [Server.MarkReady] is
	// called
	available atomic.Bool

	*cfg
}

// HealthChecker is component interface for calculating
// the current health status of a node.
type HealthChecker interface {
	// Must calculate and return current status of the node in NeoFS network map.
	//
	// If status can not be calculated for any reason,
	// control.netmapStatus_STATUS_UNDEFINED should be returned.
	NetmapStatus() control.NetmapStatus

	// Must calculate and return current health status of the node application.
	//
	// If status can not be calculated for any reason,
	// control.HealthStatus_HEALTH_STATUS_UNDEFINED should be returned.
	HealthStatus() control.HealthStatus
}

// NodeState is an interface of storage node network state.
type NodeState interface {
	// SetNetmapStatus switches the storage node to the given network status.
	//
	// If status is control.NetmapStatus_MAINTENANCE and maintenance is allowed
	// in the network settings, the node additionally starts local maintenance.
	SetNetmapStatus(st control.NetmapStatus) error

	// ForceMaintenance works like SetNetmapStatus(control.NetmapStatus_MAINTENANCE)
	// but starts local maintenance regardless of the network settings.
	ForceMaintenance() error
}

// Option of the Server's constructor.
type Option func(*cfg)

type cfg struct {
	key *ecdsa.PrivateKey

	allowedKeys [][]byte

	healthChecker HealthChecker

	netMapSrc netmap.Source

	cnrSrc container.Source

	replicator *replicator.Replicator

	nodeState NodeState

	treeService TreeService

	storage *engine.StorageEngine
}

// New creates, initializes and returns new Server instance.
// Must be marked as available with [Server.MarkReady] when all the
// components for serving are ready. Before [Server.MarkReady] call
// only health checks are available.
func New(key *ecdsa.PrivateKey, authorizedKeys [][]byte, healthChecker HealthChecker) *Server {
	cfg := &cfg{
		key:           key,
		allowedKeys:   authorizedKeys,
		healthChecker: healthChecker,
	}

	return &Server{
		cfg: cfg,
	}
}

// MarkReady marks server available. Before this call none of the other calls
// are available except for the health checks.
func (s *Server) MarkReady(e *engine.StorageEngine, nm netmap.Source, c container.Source, r *replicator.Replicator, st NodeState, tr TreeService) {
	panicOnNil := func(name string, service any) {
		if service == nil {
			panic(fmt.Sprintf("'%s' is nil", name))
		}
	}

	panicOnNil("storage engine", e)
	panicOnNil("netmap source", nm)
	panicOnNil("container source", c)
	panicOnNil("replicator", r)
	panicOnNil("node state", st)
	panicOnNil("tree service", st)

	s.storage = e
	s.netMapSrc = nm
	s.cnrSrc = c
	s.replicator = r
	s.nodeState = st
	s.treeService = tr

	s.available.Store(true)
}
