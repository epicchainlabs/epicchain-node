package alphabet

import (
	"github.com/epicchainlabs/epicchain-node/pkg/morph/event"
	netmapEvent "github.com/epicchainlabs/epicchain-node/pkg/morph/event/netmap"
	"go.uber.org/zap"
)

func (ap *Processor) HandleGasEmission(ev event.Event) {
	ne := ev.(netmapEvent.NewEpoch)

	ap.log.Info("gas emission", zap.Uint64("epoch", ne.EpochNumber()))

	// send event to the worker pool

	err := ap.pool.Submit(func() { ap.processEmit() })
	if err != nil {
		// there system can be moved into controlled degradation stage
		ap.log.Warn("alphabet processor worker pool drained",
			zap.Int("capacity", ap.pool.Cap()))
	}
}
