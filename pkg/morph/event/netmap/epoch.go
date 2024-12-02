package netmap

import (
	"fmt"

	"github.com/epicchainlabs/epicchain-go/pkg/core/state"
	"github.com/epicchainlabs/epicchain-go/pkg/util"
	"github.com/epicchainlabs/epicchain-node/pkg/morph/client"
	"github.com/epicchainlabs/epicchain-node/pkg/morph/event"
)

// NewEpoch is a new epoch Neo:Morph event.
type NewEpoch struct {
	num uint64

	// txHash is used in notary environmental
	// for calculating unique but same for
	// all notification receivers values.
	txHash util.Uint256
}

// MorphEvent implements Neo:Morph Event interface.
func (NewEpoch) MorphEvent() {}

// EpochNumber returns new epoch number.
func (s NewEpoch) EpochNumber() uint64 {
	return s.num
}

// TxHash returns hash of the TX with new epoch
// notification.
func (s NewEpoch) TxHash() util.Uint256 {
	return s.txHash
}

// ParseNewEpoch is a parser of new epoch notification event.
//
// Result is type of NewEpoch.
func ParseNewEpoch(e *state.ContainedNotificationEvent) (event.Event, error) {
	params, err := event.ParseStackArray(e)
	if err != nil {
		return nil, fmt.Errorf("could not parse stack items from notify event: %w", err)
	}

	if ln := len(params); ln != 1 {
		return nil, event.WrongNumberOfParameters(1, ln)
	}

	prmEpochNum, err := client.IntFromStackItem(params[0])
	if err != nil {
		return nil, fmt.Errorf("could not get integer epoch number: %w", err)
	}

	return NewEpoch{
		num:    uint64(prmEpochNum),
		txHash: e.Container,
	}, nil
}
