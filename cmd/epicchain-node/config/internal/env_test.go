package internal_test

import (
	"testing"

	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-node/config/internal"
	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	require.Equal(t,
		"NEOFS_SECTION_PARAMETER",
		internal.Env("section", "parameter"),
	)
}
