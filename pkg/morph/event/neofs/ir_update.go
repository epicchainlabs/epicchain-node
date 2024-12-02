package neofs

import (
	"crypto/elliptic"
	"fmt"

	"github.com/epicchainlabs/epicchain-go/pkg/crypto/keys"
	"github.com/epicchainlabs/epicchain-go/pkg/vm/stackitem"
	"github.com/epicchainlabs/epicchain-node/pkg/morph/client"
	"github.com/epicchainlabs/epicchain-node/pkg/morph/event"
)

type UpdateInnerRing struct {
	keys []*keys.PublicKey
}

// MorphEvent implements Neo:Morph Event interface.
func (UpdateInnerRing) MorphEvent() {}

func (u UpdateInnerRing) Keys() []*keys.PublicKey { return u.keys }

func ParseUpdateInnerRing(params []stackitem.Item) (event.Event, error) {
	var (
		ev  UpdateInnerRing
		err error
	)

	if ln := len(params); ln != 1 {
		return nil, event.WrongNumberOfParameters(1, ln)
	}

	// parse keys
	irKeys, err := client.ArrayFromStackItem(params[0])
	if err != nil {
		return nil, fmt.Errorf("could not get updated inner ring keys: %w", err)
	}

	ev.keys = make([]*keys.PublicKey, 0, len(irKeys))
	for i := range irKeys {
		rawKey, err := client.BytesFromStackItem(irKeys[i])
		if err != nil {
			return nil, fmt.Errorf("could not get updated inner ring public key: %w", err)
		}

		key, err := keys.NewPublicKeyFromBytes(rawKey, elliptic.P256())
		if err != nil {
			return nil, fmt.Errorf("could not parse updated inner ring public key: %w", err)
		}

		ev.keys = append(ev.keys, key)
	}

	return ev, nil
}
