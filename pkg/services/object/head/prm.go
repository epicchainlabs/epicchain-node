package headsvc

import (
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
)

type Prm struct {
	addr oid.Address
}

func (p *Prm) WithAddress(v oid.Address) *Prm {
	if p != nil {
		p.addr = v
	}

	return p
}
