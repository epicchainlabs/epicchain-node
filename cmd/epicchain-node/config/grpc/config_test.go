package grpcconfig

import (
	"testing"

	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config"
	configtest "github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config/test"
	"github.com/stretchr/testify/require"
)

func TestGRPCSection(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		require.Panics(t, func() {
			IterateEndpoints(configtest.EmptyConfig(), nil)
		})
	})

	const path = "../../../../config/example/node"

	var fileConfigTest = func(c *config.Config) {
		num := 0

		IterateEndpoints(c, func(sc *Config) {
			defer func() {
				num++
			}()

			tls := sc.TLS()

			switch num {
			case 0:
				require.Equal(t, "s01.neofs.devenv:8080", sc.Endpoint())
				require.Equal(t, 1, sc.ConnectionLimit())

				require.NotNil(t, tls)
				require.Equal(t, "/path/to/cert", tls.CertificateFile())
				require.Equal(t, "/path/to/key", tls.KeyFile())
			case 1:
				require.Equal(t, "s02.neofs.devenv:8080", sc.Endpoint())
				require.Equal(t, 0, sc.ConnectionLimit())
				require.Nil(t, tls)
			case 2:
				require.Equal(t, "s03.neofs.devenv:8080", sc.Endpoint())
				require.Equal(t, 0, sc.ConnectionLimit())
				require.Nil(t, tls)
			}
		})

		require.Equal(t, 3, num)
	}

	configtest.ForEachFileType(path, fileConfigTest)

	t.Run("ENV", func(t *testing.T) {
		configtest.ForEnvFileType(path, fileConfigTest)
	})
}
