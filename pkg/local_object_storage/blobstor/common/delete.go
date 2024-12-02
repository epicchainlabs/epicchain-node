package common

import (
	oid "github.com/epicchainlabs/neofs-sdk-go/object/id"
)

// DeletePrm groups the parameters of Delete operation.
type DeletePrm struct {
	Address   oid.Address
	StorageID []byte
}

// DeleteRes groups the resulting values of Delete operation.
type DeleteRes struct{}
