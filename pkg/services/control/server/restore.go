package control

import (
	"context"

	"github.com/epicchainlabs/epicchain-node/pkg/local_object_storage/shard"
	"github.com/epicchainlabs/epicchain-node/pkg/services/control"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) RestoreShard(_ context.Context, req *control.RestoreShardRequest) (*control.RestoreShardResponse, error) {
	err := s.isValidRequest(req)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// check availability
	err = s.ready()
	if err != nil {
		return nil, err
	}

	shardID := shard.NewIDFromBytes(req.GetBody().GetShard_ID())

	var prm shard.RestorePrm
	prm.WithPath(req.GetBody().GetFilepath())
	prm.WithIgnoreErrors(req.GetBody().GetIgnoreErrors())

	err = s.storage.RestoreShard(shardID, prm)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := new(control.RestoreShardResponse)
	resp.SetBody(new(control.RestoreShardResponse_Body))

	err = SignMessage(s.key, resp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}
