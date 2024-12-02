package acl

import (
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/acl/basic"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/acl/extended"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "acl",
	Short: "Operations with Access Control Lists",
}

func init() {
	Cmd.AddCommand(extended.Cmd)
	Cmd.AddCommand(basic.Cmd)
}
