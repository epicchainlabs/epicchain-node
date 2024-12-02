package main

import (
	"context"
	"fmt"
	"time"

	"github.com/epicchainlabs/neofs-api-go/v2/session"
	sessionGRPC "github.com/epicchainlabs/neofs-api-go/v2/session/grpc"
	nodeconfig "github.com/epicchainlabs/neofs-node/cmd/neofs-node/config/node"
	"github.com/epicchainlabs/neofs-node/pkg/morph/event"
	"github.com/epicchainlabs/neofs-node/pkg/morph/event/netmap"
	sessionTransportGRPC "github.com/epicchainlabs/neofs-node/pkg/network/transport/session/grpc"
	sessionSvc "github.com/epicchainlabs/neofs-node/pkg/services/session"
	"github.com/epicchainlabs/neofs-node/pkg/services/session/storage"
	"github.com/epicchainlabs/neofs-node/pkg/services/session/storage/persistent"
	"github.com/epicchainlabs/neofs-node/pkg/services/session/storage/temporary"
	"github.com/epicchainlabs/neofs-sdk-go/user"
)

type sessionStorage interface {
	Create(ctx context.Context, body *session.CreateRequestBody) (*session.CreateResponseBody, error)
	Get(ownerID user.ID, tokenID []byte) *storage.PrivateToken
	RemoveOld(epoch uint64)

	Close() error
}

func initSessionService(c *cfg) {
	if persistentSessionPath := nodeconfig.PersistentSessions(c.cfgReader).Path(); persistentSessionPath != "" {
		persisessions, err := persistent.NewTokenStore(persistentSessionPath,
			persistent.WithLogger(c.log),
			persistent.WithTimeout(time.Second),
			persistent.WithEncryptionKey(&c.key.PrivateKey),
		)
		if err != nil {
			panic(fmt.Errorf("could not create persistent session token storage: %w", err))
		}

		c.privateTokenStore = persisessions
	} else {
		c.privateTokenStore = temporary.NewTokenStore()
	}

	c.onShutdown(func() {
		_ = c.privateTokenStore.Close()
	})

	addNewEpochNotificationHandler(c, func(ev event.Event) {
		c.privateTokenStore.RemoveOld(ev.(netmap.NewEpoch).EpochNumber())
	})

	server := sessionTransportGRPC.New(
		sessionSvc.NewSignService(
			&c.key.PrivateKey,
			sessionSvc.NewResponseService(
				sessionSvc.NewExecutionService(c.privateTokenStore, c.log),
				c.respSvc,
			),
		),
	)

	for _, srv := range c.cfgGRPC.servers {
		sessionGRPC.RegisterSessionServiceServer(srv, server)
	}
}
