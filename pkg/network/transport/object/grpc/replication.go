package object

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	objectv2 "github.com/epicchainlabs/neofs-api-go/v2/object"
	objectGRPC "github.com/epicchainlabs/neofs-api-go/v2/object/grpc"
	refsv2 "github.com/epicchainlabs/neofs-api-go/v2/refs"
	refs "github.com/epicchainlabs/neofs-api-go/v2/refs/grpc"
	status "github.com/epicchainlabs/neofs-api-go/v2/status/grpc"
	apistatus "github.com/epicchainlabs/epicchain-sdk-go/client/status"
	cid "github.com/epicchainlabs/epicchain-sdk-go/container/id"
	neofscrypto "github.com/epicchainlabs/epicchain-sdk-go/crypto"
	neofsecdsa "github.com/epicchainlabs/epicchain-sdk-go/crypto/ecdsa"
	"github.com/epicchainlabs/epicchain-sdk-go/object"
)

// Replicate serves neo.fs.v2.object.ObjectService/Replicate RPC.
func (s *Server) Replicate(_ context.Context, req *objectGRPC.ReplicateRequest) (*objectGRPC.ReplicateResponse, error) {
	if req.Object == nil {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code: codeInternal, Message: "binary object field is missing/empty",
		}}, nil
	}

	if req.Object.ObjectId == nil || len(req.Object.ObjectId.Value) == 0 {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code: codeInternal, Message: "ID field is missing/empty in the object field",
		}}, nil
	}

	if req.Signature == nil {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code: codeInternal, Message: "missing object signature field",
		}}, nil
	}

	if len(req.Signature.Key) == 0 {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code: codeInternal, Message: "public key field is missing/empty in the object signature field",
		}}, nil
	}

	if len(req.Signature.Sign) == 0 {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code: codeInternal, Message: "signature value is missing/empty in the object signature field",
		}}, nil
	}

	switch scheme := req.Signature.Scheme; scheme {
	default:
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code:    codeInternal,
			Message: "unsupported scheme in the object signature field",
		}}, nil
	case
		refs.SignatureScheme_ECDSA_SHA512,
		refs.SignatureScheme_ECDSA_RFC6979_SHA256,
		refs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT:
	}

	hdr := req.Object.GetHeader()
	if hdr == nil {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code:    codeInternal,
			Message: "missing header field in the object field",
		}}, nil
	}

	gCnrMsg := hdr.GetContainerId()
	if gCnrMsg == nil {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code:    codeInternal,
			Message: "missing container ID field in the object header field",
		}}, nil
	}

	var cnr cid.ID
	var cnrMsg refsv2.ContainerID
	err := cnrMsg.FromGRPCMessage(gCnrMsg)
	if err == nil {
		err = cnr.ReadFromV2(cnrMsg)
	}
	if err != nil {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code:    codeInternal,
			Message: fmt.Sprintf("invalid container ID in the object header field: %v", err),
		}}, nil
	}

	var pubKey neofscrypto.PublicKey
	switch req.Signature.Scheme {
	// other cases already checked above
	case refs.SignatureScheme_ECDSA_SHA512:
		pubKey = new(neofsecdsa.PublicKey)
		err = pubKey.Decode(req.Signature.Key)
		if err != nil {
			return &objectGRPC.ReplicateResponse{Status: &status.Status{
				Code:    codeInternal,
				Message: "invalid ECDSA public key in the object signature field",
			}}, nil
		}
	case refs.SignatureScheme_ECDSA_RFC6979_SHA256:
		pubKey = new(neofsecdsa.PublicKeyRFC6979)
		err = pubKey.Decode(req.Signature.Key)
		if err != nil {
			return &objectGRPC.ReplicateResponse{Status: &status.Status{
				Code:    codeInternal,
				Message: "invalid ECDSA public key in the object signature field",
			}}, nil
		}
	case refs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT:
		pubKey = new(neofsecdsa.PublicKeyWalletConnect)
		err = pubKey.Decode(req.Signature.Key)
		if err != nil {
			return &objectGRPC.ReplicateResponse{Status: &status.Status{
				Code:    codeInternal,
				Message: "invalid ECDSA public key in the object signature field",
			}}, nil
		}
	}
	if !pubKey.Verify(req.Object.ObjectId.Value, req.Signature.Sign) {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code:    codeInternal,
			Message: "signature mismatch in the object signature field",
		}}, nil
	}

	var clientInCnr, serverInCnr bool
	err = s.node.ForEachContainerNodePublicKeyInLastTwoEpochs(cnr, func(pubKey []byte) bool {
		if !serverInCnr {
			serverInCnr = s.node.IsOwnPublicKey(pubKey)
		}
		if !clientInCnr {
			clientInCnr = bytes.Equal(pubKey, req.Signature.Key)
		}
		return !clientInCnr || !serverInCnr
	})
	if err != nil {
		if errors.Is(err, apistatus.ErrContainerNotFound) {
			return &objectGRPC.ReplicateResponse{Status: &status.Status{
				Code:    codeContainerNotFound,
				Message: "failed to check server's compliance to object's storage policy: object's container not found",
			}}, nil
		}

		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code:    codeInternal,
			Message: fmt.Sprintf("failed to apply object's storage policy: %v", err),
		}}, nil
	} else if !serverInCnr {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code: codeAccessDenied, Message: "server does not match the object's storage policy",
		}}, nil
	} else if !clientInCnr {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code: codeAccessDenied, Message: "client does not match the object's storage policy",
		}}, nil
	}

	// TODO(@cthulhu-rider): avoid decoding the object completely
	obj, err := objectFromMessage(req.Object)
	if err != nil {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code:    codeInternal,
			Message: fmt.Sprintf("invalid object field: %v", err),
		}}, nil
	}

	err = s.node.VerifyAndStoreObject(*obj)
	if err != nil {
		return &objectGRPC.ReplicateResponse{Status: &status.Status{
			Code:    codeInternal,
			Message: fmt.Sprintf("failed to verify and store object locally: %v", err),
		}}, nil
	}

	return new(objectGRPC.ReplicateResponse), nil
}

func objectFromMessage(gMsg *objectGRPC.Object) (*object.Object, error) {
	var msg objectv2.Object
	err := msg.FromGRPCMessage(gMsg)
	if err != nil {
		return nil, err
	}

	return object.NewFromV2(&msg), nil
}
