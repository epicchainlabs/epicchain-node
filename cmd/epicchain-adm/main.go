package main

import (
	"os"

	"github.com/epicchainlabs/epicchain-node/cmd/epicchain-adm/internal/modules"
)

func main() {
	if err := modules.Execute(); err != nil {
		os.Exit(1)
	}
}
