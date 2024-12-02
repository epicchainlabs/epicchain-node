package main

import (
	metricsconfig "github.com/epicchainlabs/neofs-node/cmd/neofs-node/config/metrics"
	httputil "github.com/epicchainlabs/neofs-node/pkg/util/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func initMetrics(c *cfg) *httputil.Server {
	if !metricsconfig.Enabled(c.cfgReader) {
		c.log.Info("prometheus is disabled")
		return nil
	}

	var prm httputil.Prm

	prm.Address = metricsconfig.Address(c.cfgReader)
	prm.Handler = promhttp.Handler()

	srv := httputil.New(prm,
		httputil.WithShutdownTimeout(
			metricsconfig.ShutdownTimeout(c.cfgReader),
		),
	)

	return srv
}
