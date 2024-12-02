package loggerconfig_test

import (
	"testing"

	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config"
	loggerconfig "github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config/logger"
	configtest "github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config/test"
	"github.com/stretchr/testify/require"
)

func TestLoggerSection_Level(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		v := loggerconfig.Level(configtest.EmptyConfig())
		require.Equal(t, loggerconfig.LevelDefault, v)
	})

	const path = "../../../../config/example/node"

	var fileConfigTest = func(c *config.Config) {
		v := loggerconfig.Level(c)
		require.Equal(t, "debug", v)
	}

	configtest.ForEachFileType(path, fileConfigTest)

	t.Run("ENV", func(t *testing.T) {
		configtest.ForEnvFileType(path, fileConfigTest)
	})
}
