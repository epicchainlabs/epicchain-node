package object

import (
	"github.com/epicchainlabs/neofs-sdk-go/object"
	oid "github.com/epicchainlabs/neofs-sdk-go/object/id"
)

// AddressWithType groups object address with its NeoFS
// object type.
type AddressWithType struct {
	Address oid.Address
	Type    object.Type
}
