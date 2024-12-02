package meta

import (
	"errors"

	"github.com/epicchainlabs/neofs-node/pkg/local_object_storage/util/logicerr"
	apistatus "github.com/epicchainlabs/neofs-sdk-go/client/status"
)

// ErrObjectIsExpired is returned when the requested object's
// epoch is less than the current one. Such objects are considered
// as removed and should not be returned from the Storage Engine.
var ErrObjectIsExpired = logicerr.New("object is expired")

// IsErrRemoved checks if error returned by Shard Exists/Get/Put method
// corresponds to removed object.
func IsErrRemoved(err error) bool {
	return errors.As(err, new(apistatus.ObjectAlreadyRemoved))
}
