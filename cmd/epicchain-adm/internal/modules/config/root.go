package config

import (
	"github.com/spf13/cobra"
)

const configPathFlag = "path"

var (
	// RootCmd is a root command of config section.
	RootCmd = &cobra.Command{
		Use:   "config",
		Short: "Section for epicchain-adm config related commands",
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize basic epicchain-adm configuration file",
		Example: `epicchain-adm config init
epicchain-adm config init --path .config/epicchain-adm.yml`,
		RunE: initConfig,
	}
)

func init() {
	RootCmd.AddCommand(initCmd)

	initCmd.Flags().String(configPathFlag, "", "Path to config (default ~/.neofs/adm/config.yml)")
}
