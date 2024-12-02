package cmd

import (
	"github.com/epicchainlabs/epicchain-node/pkg/util/autocomplete"
)

func init() {
	rootCmd.AddCommand(autocomplete.Command("epicchain-cli"))
}
