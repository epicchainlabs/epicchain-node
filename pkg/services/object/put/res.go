package putsvc

import (
	oid "github.com/epicchainlabs/neofs-sdk-go/object/id"
)

type PutResponse struct {
	id oid.ID
}

func (r *PutResponse) ObjectID() oid.ID {
	return r.id
}
