package deletesvc

import (
	putsvc "github.com/epicchainlabs/epicchain-node/pkg/services/object/put"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
)

type putSvcWrapper putsvc.Service

func (w *putSvcWrapper) put(exec *execCtx) (*oid.ID, error) {
	streamer, err := (*putsvc.Service)(w).Put(exec.context())
	if err != nil {
		return nil, err
	}

	payload := exec.tombstoneObj.Payload()

	initPrm := new(putsvc.PutInitPrm).
		WithCommonPrm(exec.commonParameters()).
		WithObject(exec.tombstoneObj.CutPayload())

	err = streamer.Init(initPrm)
	if err != nil {
		return nil, err
	}

	err = streamer.SendChunk(new(putsvc.PutChunkPrm).WithChunk(payload))
	if err != nil {
		return nil, err
	}

	r, err := streamer.Close()
	if err != nil {
		return nil, err
	}

	id := r.ObjectID()

	return &id, nil
}
