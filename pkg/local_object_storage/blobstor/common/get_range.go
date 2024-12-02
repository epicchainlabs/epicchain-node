package common

import (
	objectSDK "github.com/epicchainlabs/epicchain-sdk-go/object"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
)

type GetRangePrm struct {
	Address   oid.Address
	Range     objectSDK.Range
	StorageID []byte
}

type GetRangeRes struct {
	Data []byte
}
