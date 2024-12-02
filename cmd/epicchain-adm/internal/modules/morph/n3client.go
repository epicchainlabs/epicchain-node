package morph

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/epicchainlabs/epicchain-go/pkg/core/state"
	"github.com/epicchainlabs/epicchain-go/pkg/core/transaction"
	"github.com/epicchainlabs/epicchain-go/pkg/crypto/keys"
	"github.com/epicchainlabs/epicchain-go/pkg/neorpc/result"
	"github.com/epicchainlabs/epicchain-go/pkg/rpcclient"
	"github.com/epicchainlabs/epicchain-go/pkg/rpcclient/actor"
	"github.com/epicchainlabs/epicchain-go/pkg/rpcclient/invoker"
	"github.com/epicchainlabs/epicchain-go/pkg/smartcontract/trigger"
	"github.com/epicchainlabs/epicchain-go/pkg/util"
	"github.com/epicchainlabs/epicchain-go/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Client represents N3 client interface capable of test-invoking scripts
// and sending signed transactions to chain.
type Client interface {
	invoker.RPCInvoke

	GetBlockCount() (uint32, error)
	GetContractStateByID(int32) (*state.Contract, error)
	GetContractStateByHash(util.Uint160) (*state.Contract, error)
	GetNativeContracts() ([]state.Contract, error)
	GetApplicationLog(util.Uint256, *trigger.Type) (*result.ApplicationLog, error)
	GetVersion() (*result.Version, error)
	SendRawTransaction(*transaction.Transaction) (util.Uint256, error)
	GetCommittee() (keys.PublicKeys, error)
	CalculateNetworkFee(tx *transaction.Transaction) (int64, error)
}

type hashVUBPair struct {
	hash util.Uint256
	vub  uint32
}

type clientContext struct {
	Client          Client           // a raw neo-go client OR a local chain implementation
	CommitteeAct    *actor.Actor     // committee actor with the Global witness scope
	ReadOnlyInvoker *invoker.Invoker // R/O contract invoker, does not contain any signer
	SentTxs         []hashVUBPair
}

func getN3Client(v *viper.Viper) (*rpcclient.Client, error) {
	// number of opened connections
	// by neo-go client per one host
	const (
		maxConnsPerHost = 10
		requestTimeout  = time.Second * 10
	)

	ctx := context.Background()
	endpoint := v.GetString(endpointFlag)
	if endpoint == "" {
		return nil, errors.New("missing endpoint")
	}
	c, err := rpcclient.New(ctx, endpoint, rpcclient.Options{
		MaxConnsPerHost: maxConnsPerHost,
		RequestTimeout:  requestTimeout,
	})
	if err != nil {
		return nil, err
	}
	if err := c.Init(); err != nil {
		return nil, err
	}
	return c, nil
}

func defaultClientContext(c Client, committeeAcc *wallet.Account) (*clientContext, error) {
	commAct, err := actor.New(c, []actor.SignerAccount{{
		Signer: transaction.Signer{
			Account: committeeAcc.Contract.ScriptHash(),
			Scopes:  transaction.Global, // Used for test invocations only, safe to be this way.
		},
		Account: committeeAcc,
	}})
	if err != nil {
		return nil, err
	}

	return &clientContext{
		Client:          c,
		CommitteeAct:    commAct,
		ReadOnlyInvoker: invoker.New(c, nil),
	}, nil
}

func (c *clientContext) sendTx(tx *transaction.Transaction, cmd *cobra.Command, await bool) error {
	h, err := c.Client.SendRawTransaction(tx)
	if err != nil {
		return err
	}

	if h != tx.Hash() {
		return fmt.Errorf("sent and actual tx hashes mismatch:\n\tsent: %v\n\tactual: %v", tx.Hash().StringLE(), h.StringLE())
	}

	c.SentTxs = append(c.SentTxs, hashVUBPair{hash: h, vub: tx.ValidUntilBlock})

	if await {
		return c.awaitTx(cmd)
	}
	return nil
}
