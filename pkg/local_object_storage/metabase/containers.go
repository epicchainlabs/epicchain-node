package meta

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	cid "github.com/epicchainlabs/epicchain-sdk-go/container/id"
	"go.etcd.io/bbolt"
)

func (db *DB) Containers() (list []cid.ID, err error) {
	db.modeMtx.RLock()
	defer db.modeMtx.RUnlock()

	if db.mode.NoMetabase() {
		return nil, ErrDegradedMode
	}

	err = db.boltDB.View(func(tx *bbolt.Tx) error {
		list, err = db.containers(tx)

		return err
	})

	return list, err
}

func (db *DB) containers(tx *bbolt.Tx) ([]cid.ID, error) {
	result := make([]cid.ID, 0)
	unique := make(map[string]struct{})
	var cnr cid.ID

	err := tx.ForEach(func(name []byte, _ *bbolt.Bucket) error {
		if parseContainerID(&cnr, name, unique) {
			result = append(result, cnr)
			unique[string(name[1:bucketKeySize])] = struct{}{}
		}

		return nil
	})

	return result, err
}

func (db *DB) ContainerSize(id cid.ID) (size uint64, err error) {
	db.modeMtx.RLock()
	defer db.modeMtx.RUnlock()

	if db.mode.NoMetabase() {
		return 0, ErrDegradedMode
	}

	err = db.boltDB.View(func(tx *bbolt.Tx) error {
		size, err = db.containerSize(tx, id)

		return err
	})

	return size, err
}

func (db *DB) containerSize(tx *bbolt.Tx, id cid.ID) (uint64, error) {
	containerVolume := tx.Bucket(containerVolumeBucketName)
	key := make([]byte, cidSize)
	id.Encode(key)

	return parseContainerSize(containerVolume.Get(key)), nil
}

func resetContainerSize(tx *bbolt.Tx, cID cid.ID) error {
	containerVolume := tx.Bucket(containerVolumeBucketName)
	key := make([]byte, cidSize)
	cID.Encode(key)

	return containerVolume.Put(key, make([]byte, 8))
}

func parseContainerID(dst *cid.ID, name []byte, ignore map[string]struct{}) bool {
	if len(name) != bucketKeySize {
		return false
	}
	if _, ok := ignore[string(name[1:bucketKeySize])]; ok {
		return false
	}
	return dst.Decode(name[1:bucketKeySize]) == nil
}

func parseContainerSize(v []byte) uint64 {
	if len(v) == 0 {
		return 0
	}

	return binary.LittleEndian.Uint64(v)
}

func changeContainerSize(tx *bbolt.Tx, id cid.ID, delta uint64, increase bool) error {
	containerVolume := tx.Bucket(containerVolumeBucketName)
	key := make([]byte, cidSize)
	id.Encode(key)

	size := parseContainerSize(containerVolume.Get(key))

	if increase {
		size += delta
	} else if size > delta {
		size -= delta
	} else {
		size = 0
	}

	buf := make([]byte, 8) // consider using sync.Pool to decrease allocations
	binary.LittleEndian.PutUint64(buf, size)

	return containerVolume.Put(key, buf)
}

// DeleteContainer removes any information that the metabase has
// associated with the provided container (its objects) except
// the graveyard-related one.
func (db *DB) DeleteContainer(cID cid.ID) error {
	db.modeMtx.RLock()
	defer db.modeMtx.RUnlock()

	if db.mode.NoMetabase() {
		return ErrDegradedMode
	} else if db.mode.ReadOnly() {
		return ErrReadOnlyMode
	}

	cIDRaw := make([]byte, cidSize)
	cID.Encode(cIDRaw)
	buff := make([]byte, addressKeySize)

	return db.boltDB.Update(func(tx *bbolt.Tx) error {
		// Estimations
		bktEstimations := tx.Bucket(containerVolumeBucketName)
		err := bktEstimations.Delete(cIDRaw)
		if err != nil {
			return fmt.Errorf("estimations bucket cleanup: %w", err)
		}

		// Locked objects
		bktLocked := tx.Bucket(bucketNameLocked)
		err = bktLocked.DeleteBucket(cIDRaw)
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("locked bucket cleanup: %w", err)
		}

		// Regular objects
		err = tx.DeleteBucket(primaryBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("regular bucket cleanup: %w", err)
		}

		// Lock objects
		err = tx.DeleteBucket(bucketNameLockers(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("lockers bucket cleanup: %w", err)
		}

		// SG objects
		err = tx.DeleteBucket(storageGroupBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("storage group bucket cleanup: %w", err)
		}

		// TS objects
		err = tx.DeleteBucket(tombstoneBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("tombstone bucket cleanup: %w", err)
		}

		// Small objects
		err = tx.DeleteBucket(smallBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("small objects' bucket cleanup: %w", err)
		}

		// Root objects
		err = tx.DeleteBucket(rootBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("root object's bucket cleanup: %w", err)
		}

		// Link objects
		err = tx.DeleteBucket(linkObjectsBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("link objects' bucket cleanup: %w", err)
		}

		// indexes

		err = tx.DeleteBucket(ownerBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("owner index cleanup: %w", err)
		}

		err = tx.DeleteBucket(payloadHashBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("hash index cleanup: %w", err)
		}

		err = tx.DeleteBucket(parentBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("parent index cleanup: %w", err)
		}

		err = tx.DeleteBucket(splitBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("split id index cleanup: %w", err)
		}

		err = tx.DeleteBucket(firstObjectIDBucketName(cID, buff))
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return fmt.Errorf("first object id index cleanup: %w", err)
		}

		// Attributes index
		var keysToDelete [][]byte // see https://github.com/etcd-io/bbolt/issues/146
		c := tx.Cursor()
		bktPrefix := attributeBucketName(cID, "", buff)
		for k, _ := c.Seek(bktPrefix); k != nil && bytes.HasPrefix(k, bktPrefix); k, _ = c.Next() {
			keysToDelete = append(keysToDelete, k)
		}

		for _, k := range keysToDelete {
			err = tx.DeleteBucket(k)
			if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
				return fmt.Errorf("attributes index cleanup: %w", err)
			}
		}

		return nil
	})
}
