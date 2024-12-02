package innerring

import (
	"github.com/epicchainlabs/epicchain-node/pkg/innerring/processors/netmap"
	irlocode "github.com/epicchainlabs/epicchain-node/pkg/innerring/processors/netmap/nodevalidation/locode"
)

func (s *Server) newLocodeValidator() (netmap.NodeValidator, error) {
	return irlocode.New(), nil
}
