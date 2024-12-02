package meta

import (
	"os"
	"time"

	common "github.com/epicchainlabs/epicchain-node/cmd/epicchain-lens/internal"
	meta "github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/metabase"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
)

var (
	vAddress  string
	vPath     string
	vInputObj string
)

type epochState struct{}

func (s epochState) CurrentEpoch() uint64 {
	return 0
}

// Root contains `meta` command definition.
var Root = &cobra.Command{
	Use:   "meta",
	Short: "Operations with a metabase",
}

func init() {
	Root.AddCommand(
		listCMD,
		listGraveyardCMD,
		listGarbageCMD,
		writeObjectCMD,
		getCMD,
	)
}

func openMeta(cmd *cobra.Command, readOnly bool) *meta.DB {
	_, err := os.Stat(vPath)
	common.ExitOnErr(cmd, err)

	db := meta.New(
		meta.WithPath(vPath),
		meta.WithBoltDBOptions(&bbolt.Options{
			ReadOnly: readOnly,
			Timeout:  time.Second,
		}),
		meta.WithEpochState(epochState{}),
	)
	common.ExitOnErr(cmd, common.Errf("could not open metabase: %w", db.Open(readOnly)))

	return db
}
