package nodevalidation

import (
	"github.com/epicchainlabs/epicchain-node/pkg/innerring/processors/netmap"
	apinetmap "github.com/epicchainlabs/epicchain-sdk-go/netmap"
)

// CompositeValidator wraps `netmap.NodeValidator`s.
//
// For correct operation, CompositeValidator must be created
// using the constructor (New). After successful creation,
// the CompositeValidator is immediately ready to work through
// API.
type CompositeValidator struct {
	validators []netmap.NodeValidator
}

// New creates a new instance of the CompositeValidator.
//
// The created CompositeValidator does not require additional
// initialization and is completely ready for work.
func New(validators ...netmap.NodeValidator) *CompositeValidator {
	return &CompositeValidator{validators}
}

// Verify passes apinetmap.NodeInfo to wrapped validators.
//
// If error appears, returns it immediately.
func (c *CompositeValidator) Verify(ni apinetmap.NodeInfo) error {
	for _, v := range c.validators {
		if err := v.Verify(ni); err != nil {
			return err
		}
	}

	return nil
}
