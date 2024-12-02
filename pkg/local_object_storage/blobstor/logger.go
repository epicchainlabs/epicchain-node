package blobstor

import (
	storagelog "github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/internal/log"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
	"go.uber.org/zap"
)

const deleteOp = "DELETE"
const putOp = "PUT"

func logOp(l *zap.Logger, op string, addr oid.Address, typ string, sID []byte) {
	storagelog.Write(l,
		storagelog.AddressField(addr),
		storagelog.OpField(op),
		storagelog.StorageTypeField(typ),
		storagelog.StorageIDField(sID),
	)
}
