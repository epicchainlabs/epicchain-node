package morphconfig

import (
	"fmt"
	"time"

	"github.com/epicchainlabs/neofs-node/cmd/neofs-node/config"
)

const (
	subsection = "morph"

	// DialTimeoutDefault is a default dial timeout of morph chain client connection.
	DialTimeoutDefault = 5 * time.Second

	// CacheTTLDefault is a default value for cached values TTL.
	// It is 0, because actual default depends on block time.
	CacheTTLDefault = time.Duration(0)

	// ReconnectionRetriesNumberDefault is a default value for reconnection retries.
	ReconnectionRetriesNumberDefault = 5
	// ReconnectionRetriesDelayDefault is a default delay b/w reconnections.
	ReconnectionRetriesDelayDefault = 5 * time.Second
)

// Endpoints returns list of the values of "endpoints" config parameter
// from "morph" section.
//
// Throws panic if list is empty.
func Endpoints(c *config.Config) []string {
	endpoints := config.StringSliceSafe(c.Sub(subsection), "endpoints")
	if len(endpoints) == 0 {
		panic(fmt.Errorf("no morph chain RPC endpoints, see `morph.endpoints` section"))
	}
	return endpoints
}

// DialTimeout returns the value of "dial_timeout" config parameter
// from "morph" section.
//
// Returns DialTimeoutDefault if the value is not positive duration.
func DialTimeout(c *config.Config) time.Duration {
	v := config.DurationSafe(c.Sub(subsection), "dial_timeout")
	if v > 0 {
		return v
	}

	return DialTimeoutDefault
}

// CacheTTL returns the value of "cache_ttl" config parameter
// from "morph" section.
//
// Returns CacheTTLDefault if value is zero or invalid. Supports negative durations.
func CacheTTL(c *config.Config) time.Duration {
	res := config.DurationSafe(c.Sub(subsection), "cache_ttl")
	if res != 0 {
		return res
	}

	return CacheTTLDefault
}

// ReconnectionRetriesNumber returns the value of "reconnections_number" config
// parameter from "morph" section.
//
// Returns 0 if value is not specified.
func ReconnectionRetriesNumber(c *config.Config) int {
	res := config.Int(c.Sub(subsection), "reconnections_number")
	if res != 0 {
		return int(res)
	}

	return ReconnectionRetriesNumberDefault
}

// ReconnectionRetriesDelay returns the value of "reconnections_delay" config
// parameter from "morph" section.
//
// Returns 0 if value is not specified.
func ReconnectionRetriesDelay(c *config.Config) time.Duration {
	res := config.DurationSafe(c.Sub(subsection), "reconnections_delay")
	if res != 0 {
		return res
	}

	return ReconnectionRetriesDelayDefault
}
