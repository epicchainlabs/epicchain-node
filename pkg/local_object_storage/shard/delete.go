package shard

import (
	"errors"

	"github.com/epicchainlabs/neofs-node/pkg/local_object_storage/blobstor/common"
	meta "github.com/epicchainlabs/neofs-node/pkg/local_object_storage/metabase"
	"github.com/epicchainlabs/neofs-node/pkg/local_object_storage/writecache"
	oid "github.com/epicchainlabs/neofs-sdk-go/object/id"
	"go.uber.org/zap"
)

// DeletePrm groups the parameters of Delete operation.
type DeletePrm struct {
	addr              []oid.Address
	skipNotFoundError bool
}

// DeleteRes groups the resulting values of Delete operation.
type DeleteRes struct{}

// SetAddresses is a Delete option to set the addresses of the objects to delete.
//
// Option is required.
func (p *DeletePrm) SetAddresses(addr ...oid.Address) {
	p.addr = append(p.addr, addr...)
}

// SkipNotFoundError is a Delete option to skip errors when an already deleted
// object is being deleted.
func (p *DeletePrm) SkipNotFoundError() {
	p.skipNotFoundError = true
}

// Delete removes data from the shard's writeCache, metaBase and
// blobStor.
func (s *Shard) Delete(prm DeletePrm) (DeleteRes, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.delete(prm)
}

func (s *Shard) delete(prm DeletePrm) (DeleteRes, error) {
	if s.info.Mode.ReadOnly() {
		return DeleteRes{}, ErrReadOnlyMode
	} else if s.info.Mode.NoMetabase() {
		return DeleteRes{}, ErrDegradedMode
	}

	ln := len(prm.addr)

	smalls := make(map[oid.Address][]byte, ln)

	for i := range prm.addr {
		if s.hasWriteCache() {
			err := s.writeCache.Delete(prm.addr[i])
			if err != nil && !IsErrNotFound(err) && !errors.Is(err, writecache.ErrReadOnly) {
				s.log.Warn("can't delete object from write cache", zap.String("error", err.Error()))
			}
		}

		var sPrm meta.StorageIDPrm
		sPrm.SetAddress(prm.addr[i])

		res, err := s.metaBase.StorageID(sPrm)
		if err != nil {
			s.log.Debug("can't get storage ID from metabase",
				zap.Stringer("object", prm.addr[i]),
				zap.String("error", err.Error()))

			continue
		}

		if res.StorageID() != nil {
			smalls[prm.addr[i]] = res.StorageID()
		}
	}

	var delPrm meta.DeletePrm
	delPrm.SetAddresses(prm.addr...)

	res, err := s.metaBase.Delete(delPrm)
	if err != nil {
		return DeleteRes{}, err // stop on metabase error ?
	}

	s.decObjectCounterBy(physical, res.RawObjectsRemoved())
	s.decObjectCounterBy(logical, res.AvailableObjectsRemoved())

	var totalRemovedPayload uint64
	removedSizes := res.RemovedObjectSizes()

	for i := range prm.addr {
		removedPayload := removedSizes[i]
		totalRemovedPayload += removedPayload
		s.addToContainerSize(prm.addr[i].Container().EncodeToString(), -int64(removedPayload))
	}

	s.addToPayloadCounter(-int64(totalRemovedPayload))

	for i := range prm.addr {
		var delPrm common.DeletePrm
		delPrm.Address = prm.addr[i]
		id := smalls[prm.addr[i]]
		delPrm.StorageID = id

		_, err = s.blobStor.Delete(delPrm)
		if err != nil {
			if IsErrNotFound(err) && prm.skipNotFoundError {
				continue
			}

			s.log.Debug("can't remove object from blobStor",
				zap.Stringer("object_address", prm.addr[i]),
				zap.String("error", err.Error()))
		}
	}

	return DeleteRes{}, nil
}
