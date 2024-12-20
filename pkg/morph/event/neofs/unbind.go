package neofs

import (
	"fmt"

	"github.com/epicchainlabs/epicchain-go/pkg/core/state"
	"github.com/epicchainlabs/epicchain-node/pkg/morph/event"
)

type Unbind struct {
	bindCommon
}

func ParseUnbind(e *state.ContainedNotificationEvent) (event.Event, error) {
	var (
		ev  Unbind
		err error
	)

	params, err := event.ParseStackArray(e)
	if err != nil {
		return nil, fmt.Errorf("could not parse stack items from notify event: %w", err)
	}

	err = parseBind(&ev.bindCommon, params)
	if err != nil {
		return nil, err
	}

	ev.txHash = e.Container

	return ev, nil
}
