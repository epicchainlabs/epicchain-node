package cmd

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/internal/common"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/internal/commonflags"
	accountingCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/accounting"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/acl"
	bearerCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/bearer"
	containerCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/container"
	controlCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/control"
	netmapCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/netmap"
	objectCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/object"
	sessionCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/session"
	sgCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/storagegroup"
	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/tree"
	utilCli "github.com/epicchainlabs/epicchain-node/cmd/epicchain-cli/modules/util"
	"github.com/epicchainlabs/epicchain-node/misc"
	"github.com/epicchainlabs/epicchain-node/pkg/util/gendoc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	envPrefix = "NEOFS_CLI"
)

// Global scope flags.
var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "epicchain-cli",
	Short: "Command Line Tool to work with NeoFS",
	Long: `NeoFS CLI provides all basic interactions with NeoFS and it's services.

It contains commands for interaction with NeoFS nodes using different versions
of neofs-api and some useful utilities for compiling ACL rules from JSON
notation, managing container access through protocol gates, querying network map
and much more!`,
	Args: cobra.NoArgs,
	Run:  entryPoint,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	common.ExitOnErr(rootCmd, "", err)
}

func init() {
	cobra.OnInitialize(initConfig)

	// use stdout as default output for cmd.Print()
	rootCmd.SetOut(os.Stdout)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file (default is $HOME/.config/epicchain-cli/config.yaml)")
	rootCmd.PersistentFlags().BoolP(commonflags.Verbose, commonflags.VerboseShorthand,
		false, commonflags.VerboseUsage)

	_ = viper.BindPFlag(commonflags.Verbose, rootCmd.PersistentFlags().Lookup(commonflags.Verbose))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().Bool("version", false, "Application version and NeoFS API compatibility")

	rootCmd.AddCommand(acl.Cmd)
	rootCmd.AddCommand(bearerCli.Cmd)
	rootCmd.AddCommand(sessionCli.Cmd)
	rootCmd.AddCommand(accountingCli.Cmd)
	rootCmd.AddCommand(controlCli.Cmd)
	rootCmd.AddCommand(utilCli.Cmd)
	rootCmd.AddCommand(netmapCli.Cmd)
	rootCmd.AddCommand(objectCli.Cmd)
	rootCmd.AddCommand(sgCli.Cmd)
	rootCmd.AddCommand(containerCli.Cmd)
	rootCmd.AddCommand(tree.Cmd)
	rootCmd.AddCommand(gendoc.Command(rootCmd))
}

func entryPoint(cmd *cobra.Command, _ []string) {
	printVersion, _ := cmd.Flags().GetBool("version")
	if printVersion {
		cmd.Print(misc.BuildInfo("NeoFS CLI"))

		return
	}

	_ = cmd.Usage()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		common.ExitOnErr(rootCmd, "", err)

		// Search config in `$HOME/.config/epicchain-cli/` with name "config.yaml"
		viper.AddConfigPath(filepath.Join(home, ".config", "epicchain-cli"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		common.PrintVerbose(rootCmd, "Using config file: %s", viper.ConfigFileUsed())
	}
}
