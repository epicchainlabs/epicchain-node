package main

import (
	"os"

	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-lens/internal/meta"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-lens/internal/object"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-lens/internal/peapod"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-lens/internal/storage"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-lens/internal/writecache"
	"github.com/epicchainlabs/epicchain-node/misc"
	"github.com/epicchainlabs/epicchain-node/pkg/util/gendoc"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use:          "epicchain-lens",
	Short:        "NeoFS Storage Engine Lens",
	Long:         `NeoFS Storage Engine Lens provides tools to browse the contents of the NeoFS storage engine.`,
	RunE:         entryPoint,
	SilenceUsage: true,
}

func entryPoint(cmd *cobra.Command, _ []string) error {
	printVersion, _ := cmd.Flags().GetBool("version")
	if printVersion {
		cmd.Print(misc.BuildInfo("NeoFS Lens"))

		return nil
	}

	return cmd.Usage()
}

func init() {
	// use stdout as default output for cmd.Print()
	command.SetOut(os.Stdout)
	command.Flags().Bool("version", false, "Application version")
	command.AddCommand(
		peapod.Root,
		meta.Root,
		writecache.Root,
		storage.Root,
		object.Root,
		gendoc.Command(command),
	)
}

func main() {
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
