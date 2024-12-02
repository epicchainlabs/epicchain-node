package netmap

import (
	"github.com/epicchainlabs/neofs-node/cmd/internal/cmdprinter"
	internalclient "github.com/epicchainlabs/neofs-node/cmd/neofs-cli/internal/client"
	"github.com/epicchainlabs/neofs-node/cmd/neofs-cli/internal/common"
	"github.com/epicchainlabs/neofs-node/cmd/neofs-cli/internal/commonflags"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Request current local snapshot of the network map",
	Long:  `Request current local snapshot of the network map`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx, cancel := commonflags.GetCommandContext(cmd)
		defer cancel()

		cli := internalclient.GetSDKClientByFlag(ctx, cmd, commonflags.RPC)

		var prm internalclient.NetMapSnapshotPrm
		prm.SetClient(cli)

		res, err := internalclient.NetMapSnapshot(ctx, prm)
		common.ExitOnErr(cmd, "rpc error: %w", err)

		cmdprinter.PrettyPrintNetMap(cmd, res.NetMap())
	},
}

func initSnapshotCmd() {
	commonflags.Init(snapshotCmd)
	commonflags.InitAPI(snapshotCmd)
}
