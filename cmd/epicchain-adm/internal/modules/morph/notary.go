package morph

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/epicchainlabs/epicchain-go/cli/input"
	"github.com/epicchainlabs/epicchain-go/pkg/core/transaction"
	"github.com/epicchainlabs/epicchain-go/pkg/crypto/keys"
	"github.com/epicchainlabs/epicchain-go/pkg/encoding/address"
	"github.com/epicchainlabs/epicchain-go/pkg/rpcclient/actor"
	"github.com/epicchainlabs/epicchain-go/pkg/rpcclient/gas"
	"github.com/epicchainlabs/epicchain-go/pkg/rpcclient/notary"
	"github.com/epicchainlabs/epicchain-go/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// defaultNotaryDepositLifetime is an amount of blocks notary deposit stays valid.
// https://github.com/epicchainlabs/epicchain-go/blob/master/pkg/core/native/notary.go#L48
const defaultNotaryDepositLifetime = 5760

func depositNotary(cmd *cobra.Command, _ []string) error {
	p, err := cmd.Flags().GetString(storageWalletFlag)
	if err != nil {
		return err
	} else if p == "" {
		return fmt.Errorf("missing wallet path (use '--%s <out.json>')", storageWalletFlag)
	}

	w, err := wallet.NewWalletFromFile(p)
	if err != nil {
		return fmt.Errorf("can't open wallet: %v", err)
	}

	accHash := w.GetChangeAddress()
	if addr, err := cmd.Flags().GetString(walletAccountFlag); err == nil {
		accHash, err = address.StringToUint160(addr)
		if err != nil {
			return fmt.Errorf("invalid address: %s", addr)
		}
	}

	acc := w.GetAccount(accHash)
	if acc == nil {
		return fmt.Errorf("can't find account for %s", accHash)
	}

	prompt := fmt.Sprintf("Enter password for %s >", address.Uint160ToString(accHash))
	pass, err := input.ReadPassword(prompt)
	if err != nil {
		return fmt.Errorf("can't get password: %v", err)
	}

	err = acc.Decrypt(pass, keys.NEP2ScryptParams())
	if err != nil {
		return fmt.Errorf("can't unlock account: %v", err)
	}

	gasStr, err := cmd.Flags().GetString(refillGasAmountFlag)
	if err != nil {
		return err
	}
	gasAmount, err := parseGASAmount(gasStr)
	if err != nil {
		return err
	}

	till := int64(defaultNotaryDepositLifetime)
	tillStr, err := cmd.Flags().GetString(notaryDepositTillFlag)
	if err != nil {
		return err
	}
	if tillStr != "" {
		till, err = strconv.ParseInt(tillStr, 10, 64)
		if err != nil || till <= 0 {
			return fmt.Errorf("notary deposit lifetime must be a positive integer")
		}
	}

	c, err := getN3Client(viper.GetViper())
	if err != nil {
		return err
	}

	height, err := c.GetBlockCount()
	if err != nil {
		return fmt.Errorf("can't get current height: %v", err)
	}

	act, err := actor.New(c, []actor.SignerAccount{{
		Signer: transaction.Signer{
			Account: acc.Contract.ScriptHash(),
			Scopes:  transaction.Global,
		},
		Account: acc,
	}})
	if err != nil {
		return fmt.Errorf("could not create actor: %w", err)
	}

	gasActor := gas.New(act)
	transferData := &notary.OnNEP17PaymentData{Till: height + uint32(till)}

	txHash, vub, err := gasActor.Transfer(
		accHash,
		notary.Hash,
		big.NewInt(int64(gasAmount)),
		transferData,
	)
	if err != nil {
		return fmt.Errorf("could not send tx: %w", err)
	}

	return awaitTx(cmd, c, []hashVUBPair{{hash: txHash, vub: vub}})
}
