package storagegroup

import (
	internalclient "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/internal/client"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/internal/common"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/internal/commonflags"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/internal/key"
	objectCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/object"
	"github.com/epicchainlabs/epicchain-node/pkg/services/object_manager/storagegroup"
	cid "github.com/epicchainlabs/epicchain-sdk-go/container/id"
	"github.com/spf13/cobra"
)

var sgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storage groups in NeoFS container",
	Long:  "List storage groups in NeoFS container",
	Args:  cobra.NoArgs,
	Run:   listSG,
}

func initSGListCmd() {
	commonflags.Init(sgListCmd)

	sgListCmd.Flags().String(commonflags.CIDFlag, "", commonflags.CIDFlagUsage)
	_ = sgListCmd.MarkFlagRequired(commonflags.CIDFlag)
}

func listSG(cmd *cobra.Command, _ []string) {
	ctx, cancel := commonflags.GetCommandContext(cmd)
	defer cancel()

	var cnr cid.ID
	readCID(cmd, &cnr)

	pk := key.GetOrGenerate(cmd)

	cli := internalclient.GetSDKClientByFlag(ctx, cmd, commonflags.RPC)

	var prm internalclient.SearchObjectsPrm
	objectCli.Prepare(cmd, &prm)
	prm.SetClient(cli)
	prm.SetPrivateKey(*pk)
	prm.SetContainerID(cnr)
	prm.SetFilters(storagegroup.SearchQuery())

	res, err := internalclient.SearchObjects(ctx, prm)
	common.ExitOnErr(cmd, "rpc error: %w", err)

	ids := res.IDList()

	cmd.Printf("Found %d storage groups.\n", len(ids))

	for i := range ids {
		cmd.Println(ids[i].String())
	}
}
