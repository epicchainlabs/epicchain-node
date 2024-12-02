package main

import (
	"os"

	"github.com/epicchainlabs/neofs-node/cmd/neofs-adm/internal/modules"
)

func main() {
	if err := modules.Execute(); err != nil {
		os.Exit(1)
	}
}
