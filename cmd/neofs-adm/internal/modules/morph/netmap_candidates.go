package morph

import (
	"fmt"

	"github.com/epicchainlabs/epicchain-go/pkg/rpcclient/invoker"
	"github.com/epicchainlabs/neofs-contract/rpc/nns"
	"github.com/epicchainlabs/neofs-node/cmd/internal/cmdprinter"
	"github.com/epicchainlabs/neofs-node/pkg/morph/client/netmap"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func listNetmapCandidatesNodes(cmd *cobra.Command, _ []string) error {
	c, err := getN3Client(viper.GetViper())
	if err != nil {
		return err
	}

	inv := invoker.New(c, nil)

	nnsReader, err := nns.NewInferredReader(c, inv)
	if err != nil {
		return fmt.Errorf("can't find NNS contract: %w", err)
	}

	nmHash, err := nnsReader.ResolveFSContract(nns.NameNetmap)
	if err != nil {
		return fmt.Errorf("can't get netmap contract hash: %w", err)
	}

	res, err := inv.Call(nmHash, "netmapCandidates")
	if err != nil {
		return fmt.Errorf("can't fetch list of network config keys from the netmap contract: %w", err)
	}
	if res.State != "HALT" {
		return fmt.Errorf("netmap contract returned unexpected exception: %s", res.FaultException)
	}

	nm, err := netmap.DecodeNetMap(res.Stack)

	if err != nil {
		return fmt.Errorf("unable to decode netmap: %w", err)
	}
	nodes := nm.Nodes()
	for i := range nodes {
		cmdprinter.PrettyPrintNodeInfo(cmd, nodes[i], i, "", false)
	}
	return nil
}
