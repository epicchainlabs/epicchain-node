package shard

import (
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/blobstor"
	meta "github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/metabase"
	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/writecache"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
)

// ObjectStatus represents the status of an object in a storage system. It contains
// information about the object's status in various sub-components such as Blob storage,
// Metabase, and Writecache. Additionally, it includes a slice of errors that may have
// occurred at the object level.
type ObjectStatus struct {
	Blob       blobstor.ObjectStatus
	Metabase   meta.ObjectStatus
	Writecache writecache.ObjectStatus
	Errors     []error
}

// ObjectStatus returns the status of the object in the Shard. It contains status
// of the object in Blob storage, Metabase and Writecache.
func (s *Shard) ObjectStatus(address oid.Address) (ObjectStatus, error) {
	var res ObjectStatus
	var err error
	res.Blob, err = s.blobStor.ObjectStatus(address)
	if len(res.Blob.Substorages) != 0 {
		res.Errors = append(res.Errors, err)
		res.Metabase, err = s.metaBase.ObjectStatus(address)
		res.Errors = append(res.Errors, err)
		if s.hasWriteCache() {
			res.Writecache, err = s.writeCache.ObjectStatus(address)
			res.Errors = append(res.Errors, err)
		}
	}
	return res, nil
}
