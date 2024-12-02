package deletesvc

import (
	"context"
	"strconv"

	"github.com/epicchainlabs/epicchain-node/pkg/services/object/util"
	cid "github.com/epicchainlabs/epicchain-sdk-go/container/id"
	"github.com/epicchainlabs/epicchain-sdk-go/object"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
	"go.uber.org/zap"
)

type statusError struct {
	status int
	err    error
}

type execCtx struct {
	svc *Service

	ctx context.Context

	prm Prm

	statusError

	log *zap.Logger

	tombstone *object.Tombstone

	tombstoneObj *object.Object
}

const (
	statusUndefined int = iota
	statusOK
)

func (exec *execCtx) setLogger(l *zap.Logger) {
	exec.log = l.With(
		zap.String("request", "DELETE"),
		zap.Stringer("address", exec.address()),
		zap.Bool("local", exec.isLocal()),
		zap.Bool("with session", exec.prm.common.SessionToken() != nil),
		zap.Bool("with bearer", exec.prm.common.BearerToken() != nil),
	)
}

func (exec execCtx) context() context.Context {
	return exec.ctx
}

func (exec execCtx) isLocal() bool {
	return exec.prm.common.LocalOnly()
}

func (exec *execCtx) address() oid.Address {
	return exec.prm.addr
}

func (exec *execCtx) containerID() cid.ID {
	return exec.prm.addr.Container()
}

func (exec *execCtx) commonParameters() *util.CommonPrm {
	return exec.prm.common
}

func (exec *execCtx) newAddress(id oid.ID) oid.Address {
	var a oid.Address
	a.SetObject(id)
	a.SetContainer(exec.containerID())

	return a
}

func (exec *execCtx) addMembers(incoming []oid.ID) {
	members := exec.tombstone.Members()

	for i := range members {
		for j := 0; j < len(incoming); j++ { // don't use range, slice mutates in body
			if members[i].Equals(incoming[j]) {
				incoming = append(incoming[:j], incoming[j+1:]...)
				j--
			}
		}
	}

	exec.tombstone.SetMembers(append(members, incoming...))
}

func (exec *execCtx) initTombstoneObject() bool {
	payload, err := exec.tombstone.Marshal()
	if err != nil {
		exec.status = statusUndefined
		exec.err = err

		exec.log.Debug("could not marshal tombstone structure",
			zap.String("error", err.Error()),
		)

		return false
	}

	exec.tombstoneObj = object.New()
	exec.tombstoneObj.SetContainerID(exec.containerID())
	exec.tombstoneObj.SetType(object.TypeTombstone)
	exec.tombstoneObj.SetPayload(payload)

	tokenSession := exec.commonParameters().SessionToken()
	if tokenSession != nil {
		issuer := tokenSession.Issuer()
		exec.tombstoneObj.SetOwnerID(&issuer)
	} else {
		// make local node a tombstone object owner
		localUser := exec.svc.netInfo.LocalNodeID()
		exec.tombstoneObj.SetOwnerID(&localUser)
	}

	var a object.Attribute
	a.SetKey(object.AttributeExpirationEpoch)
	a.SetValue(strconv.FormatUint(exec.tombstone.ExpirationEpoch(), 10))

	exec.tombstoneObj.SetAttributes(a)

	return true
}

func (exec *execCtx) saveTombstone() bool {
	id, err := exec.svc.placer.put(exec)

	switch {
	default:
		exec.status = statusUndefined
		exec.err = err

		exec.log.Debug("could not save the tombstone",
			zap.String("error", err.Error()),
		)

		return false
	case err == nil:
		exec.status = statusOK
		exec.err = nil

		exec.prm.tombAddrWriter.
			SetAddress(exec.newAddress(*id))
	}

	return true
}
