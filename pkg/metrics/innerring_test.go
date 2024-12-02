package metrics_test

import (
	"testing"

	"github.com/epicchainlabs/neofs-node/pkg/metrics"
	"github.com/stretchr/testify/require"
)

func TestNewInnerRingMetrics(t *testing.T) {
	require.NotPanics(t, func() {
		_ = metrics.NewInnerRingMetrics("any_version")
	})
}
