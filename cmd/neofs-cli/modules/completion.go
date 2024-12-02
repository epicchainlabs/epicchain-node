package cmd

import (
	"github.com/epicchainlabs/neofs-node/pkg/util/autocomplete"
)

func init() {
	rootCmd.AddCommand(autocomplete.Command("neofs-cli"))
}
