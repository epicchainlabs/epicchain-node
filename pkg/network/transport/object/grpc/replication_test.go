package object_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"

	objectV2 "github.com/epicchainlabs/neofs-api-go/v2/object"
	objectgrpc "github.com/epicchainlabs/neofs-api-go/v2/object/grpc"
	refs "github.com/epicchainlabs/neofs-api-go/v2/refs/grpc"
	. "github.com/epicchainlabs/epicchain-node/pkg/network/transport/object/grpc"
	objectSvc "github.com/epicchainlabs/epicchain-node/pkg/services/object"
	cid "github.com/epicchainlabs/epicchain-sdk-go/container/id"
	cidtest "github.com/epicchainlabs/epicchain-sdk-go/container/id/test"
	neofscrypto "github.com/epicchainlabs/epicchain-sdk-go/crypto"
	neofsecdsa "github.com/epicchainlabs/epicchain-sdk-go/crypto/ecdsa"
	"github.com/epicchainlabs/epicchain-sdk-go/crypto/test"
	"github.com/epicchainlabs/epicchain-sdk-go/object"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
	oidtest "github.com/epicchainlabs/epicchain-sdk-go/object/id/test"
	objecttest "github.com/epicchainlabs/epicchain-sdk-go/object/test"
	"github.com/stretchr/testify/require"
)

func randECDSAPrivateKey(tb testing.TB) *ecdsa.PrivateKey {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(tb, err)
	return k
}

type noCallObjectService struct{}

func (x noCallObjectService) Get(*objectV2.GetRequest, objectSvc.GetObjectStream) error {
	panic("must not be called")
}

func (x noCallObjectService) Put(context.Context) (objectSvc.PutObjectStream, error) {
	panic("must not be called")
}

func (x noCallObjectService) Head(context.Context, *objectV2.HeadRequest) (*objectV2.HeadResponse, error) {
	panic("must not be called")
}

func (x noCallObjectService) Search(*objectV2.SearchRequest, objectSvc.SearchStream) error {
	panic("must not be called")
}

func (x noCallObjectService) Delete(context.Context, *objectV2.DeleteRequest) (*objectV2.DeleteResponse, error) {
	panic("must not be called")
}

func (x noCallObjectService) GetRange(*objectV2.GetRangeRequest, objectSvc.GetObjectRangeStream) error {
	panic("must not be called")
}

func (x noCallObjectService) GetRangeHash(context.Context, *objectV2.GetRangeHashRequest) (*objectV2.GetRangeHashResponse, error) {
	panic("must not be called")
}

type noCallTestNode struct{}

func (x *noCallTestNode) ForEachContainerNodePublicKeyInLastTwoEpochs(cid.ID, func([]byte) bool) error {
	panic("must not be called")
}

func (x *noCallTestNode) IsOwnPublicKey([]byte) bool {
	panic("must not be called")
}

func (x *noCallTestNode) VerifyAndStoreObject(object.Object) error {
	panic("must not be called")
}

type testNode struct {
	tb testing.TB

	// server state
	serverPubKey []byte

	// request data
	clientPubKey []byte
	cnr          cid.ID
	obj          *objectgrpc.Object

	// return
	cnrErr           error
	clientOutsideCnr bool
	serverOutsideCnr bool

	storeErr error
}

func newTestNode(tb testing.TB, serverPubKey, clientPubKey []byte, cnr cid.ID, obj *objectgrpc.Object) *testNode {
	return &testNode{
		tb:           tb,
		serverPubKey: serverPubKey,
		clientPubKey: clientPubKey,
		cnr:          cnr,
		obj:          obj,
	}
}

func (x *testNode) ForEachContainerNodePublicKeyInLastTwoEpochs(cnr cid.ID, f func(pubKey []byte) bool) error {
	require.True(x.tb, cnr.Equals(x.cnr))
	require.NotNil(x.tb, f)
	if x.cnrErr != nil {
		return x.cnrErr
	}
	if !x.clientOutsideCnr && !f(x.clientPubKey) {
		return nil
	}
	if !x.serverOutsideCnr && !f(x.serverPubKey) {
		return nil
	}
	return nil
}

func (x *testNode) IsOwnPublicKey(pubKey []byte) bool { return bytes.Equal(x.serverPubKey, pubKey) }

func (x *testNode) VerifyAndStoreObject(obj object.Object) error {
	require.Equal(x.tb, x.obj, obj.ToV2().ToGRPCMessage().(*objectgrpc.Object))
	return x.storeErr
}

func anyValidRequest(tb testing.TB, signer neofscrypto.Signer, cnr cid.ID, objID oid.ID) *objectgrpc.ReplicateRequest {
	obj := objecttest.Object(tb)
	obj.SetContainerID(cnr)
	obj.SetID(objID)

	sig, err := signer.Sign(objID[:])
	require.NoError(tb, err)

	req := &objectgrpc.ReplicateRequest{
		Object: obj.ToV2().ToGRPCMessage().(*objectgrpc.Object),
		Signature: &refs.Signature{
			Key:  neofscrypto.PublicKeyBytes(signer.Public()),
			Sign: sig,
		},
	}

	switch signer.Scheme() {
	default:
		tb.Fatalf("unsupported scheme %v", signer.Scheme())
	case neofscrypto.ECDSA_SHA512:
		req.Signature.Scheme = refs.SignatureScheme_ECDSA_SHA512
	case neofscrypto.ECDSA_DETERMINISTIC_SHA256:
		req.Signature.Scheme = refs.SignatureScheme_ECDSA_RFC6979_SHA256
	case neofscrypto.ECDSA_WALLETCONNECT:
		req.Signature.Scheme = refs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT
	}

	return req
}

func TestServer_Replicate(t *testing.T) {
	var noCallNode noCallTestNode
	var noCallObjSvc noCallObjectService
	noCallSrv := New(noCallObjSvc, &noCallNode)
	clientSigner := test.RandomSigner(t)
	clientPubKey := neofscrypto.PublicKeyBytes(clientSigner.Public())
	serverPubKey := neofscrypto.PublicKeyBytes(test.RandomSigner(t).Public())
	cnr := cidtest.ID()
	objID := oidtest.ID()
	req := anyValidRequest(t, clientSigner, cnr, objID)

	t.Run("invalid/unsupported signature format", func(t *testing.T) {
		// note: verification is tested separately
		for _, tc := range []struct {
			name         string
			fSig         func() *refs.Signature
			expectedCode uint32
			expectedMsg  string
		}{
			{
				name:         "missing object signature field",
				fSig:         func() *refs.Signature { return nil },
				expectedCode: 1024,
				expectedMsg:  "missing object signature field",
			},
			{
				name: "missing public key field in the signature field",
				fSig: func() *refs.Signature {
					return &refs.Signature{
						Key:    nil,
						Sign:   []byte("any non-empty"),
						Scheme: refs.SignatureScheme_ECDSA_SHA512, // any supported
					}
				},
				expectedCode: 1024,
				expectedMsg:  "public key field is missing/empty in the object signature field",
			},
			{
				name: "missing value field in the signature field",
				fSig: func() *refs.Signature {
					return &refs.Signature{
						Key:    []byte("any non-empty"),
						Sign:   []byte{},
						Scheme: refs.SignatureScheme_ECDSA_SHA512, // any supported
					}
				},
				expectedCode: 1024,
				expectedMsg:  "signature value is missing/empty in the object signature field",
			},
			{
				name: "unsupported scheme in the signature field",
				fSig: func() *refs.Signature {
					return &refs.Signature{
						Key:    []byte("any non-empty"),
						Sign:   []byte("any non-empty"),
						Scheme: 3,
					}
				},
				expectedCode: 1024,
				expectedMsg:  "unsupported scheme in the object signature field",
			},
		} {
			req := anyValidRequest(t, test.RandomSigner(t), cidtest.ID(), oidtest.ID())
			req.Signature = tc.fSig()
			resp, err := noCallSrv.Replicate(context.Background(), req)
			require.NoError(t, err, tc.name)
			require.EqualValues(t, tc.expectedCode, resp.GetStatus().GetCode(), tc.name)
			require.Equal(t, tc.expectedMsg, resp.GetStatus().GetMessage(), tc.name)
		}
	})

	t.Run("signature verification failure", func(t *testing.T) {
		// note: common format is tested separately
		for _, tc := range []struct {
			name         string
			fSig         func(bObj []byte) *refs.Signature
			expectedCode uint32
			expectedMsg  string
		}{
			{
				name: "ECDSA SHA-512: invalid public key",
				fSig: func(_ []byte) *refs.Signature {
					return &refs.Signature{
						Key:    []byte("not ECDSA key"),
						Sign:   []byte("any non-empty"),
						Scheme: refs.SignatureScheme_ECDSA_SHA512,
					}
				},
				expectedCode: 1024,
				expectedMsg:  "invalid ECDSA public key in the object signature field",
			},
			{
				name: "ECDSA SHA-512: signature mismatch",
				fSig: func(bObj []byte) *refs.Signature {
					return &refs.Signature{
						Key:    neofscrypto.PublicKeyBytes((*neofsecdsa.Signer)(randECDSAPrivateKey(t)).Public()),
						Sign:   []byte("definitely invalid"),
						Scheme: refs.SignatureScheme_ECDSA_SHA512,
					}
				},
				expectedCode: 1024,
				expectedMsg:  "signature mismatch in the object signature field",
			},
			{
				name: "ECDSA SHA-256 deterministic: invalid public key",
				fSig: func(_ []byte) *refs.Signature {
					return &refs.Signature{
						Key:    []byte("not ECDSA key"),
						Sign:   []byte("any non-empty"),
						Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256,
					}
				},
				expectedCode: 1024,
				expectedMsg:  "invalid ECDSA public key in the object signature field",
			},
			{
				name: "ECDSA SHA-256 deterministic: signature mismatch",
				fSig: func(bObj []byte) *refs.Signature {
					return &refs.Signature{
						Key:    neofscrypto.PublicKeyBytes((*neofsecdsa.SignerRFC6979)(randECDSAPrivateKey(t)).Public()),
						Sign:   []byte("definitely invalid"),
						Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256,
					}
				},
				expectedCode: 1024,
				expectedMsg:  "signature mismatch in the object signature field",
			},
			{
				name: "ECDSA SHA-256 WalletConnect: invalid public key",
				fSig: func(_ []byte) *refs.Signature {
					return &refs.Signature{
						Key:    []byte("not ECDSA key"),
						Sign:   []byte("any non-empty"),
						Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT,
					}
				},
				expectedCode: 1024,
				expectedMsg:  "invalid ECDSA public key in the object signature field",
			},
			{
				name: "ECDSA SHA-256 WalletConnect: signature mismatch",
				fSig: func(bObj []byte) *refs.Signature {
					return &refs.Signature{
						Key:    neofscrypto.PublicKeyBytes((*neofsecdsa.SignerWalletConnect)(randECDSAPrivateKey(t)).Public()),
						Sign:   []byte("definitely invalid"),
						Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT,
					}
				},
				expectedCode: 1024,
				expectedMsg:  "signature mismatch in the object signature field",
			},
		} {
			obj := objecttest.Object(t)
			bObj, err := obj.Marshal()
			require.NoError(t, err)

			resp, err := noCallSrv.Replicate(context.Background(), &objectgrpc.ReplicateRequest{
				Object:    obj.ToV2().ToGRPCMessage().(*objectgrpc.Object),
				Signature: tc.fSig(bObj),
			})
			require.NoError(t, err, tc.name)
			require.EqualValues(t, tc.expectedCode, resp.GetStatus().GetCode(), tc.name)
			require.Equal(t, tc.expectedMsg, resp.GetStatus().GetMessage(), tc.name)
		}
	})

	t.Run("apply storage policy failure", func(t *testing.T) {
		node := newTestNode(t, serverPubKey, clientPubKey, cnr, req.Object)
		srv := New(noCallObjSvc, node)

		node.cnrErr = errors.New("any error")

		resp, err := srv.Replicate(context.Background(), req)
		require.NoError(t, err)
		require.EqualValues(t, 1024, resp.GetStatus().GetCode())
		require.Equal(t, "failed to apply object's storage policy: any error", resp.GetStatus().GetMessage())
	})

	t.Run("client or server mismatches object's storage policy", func(t *testing.T) {
		node := newTestNode(t, serverPubKey, clientPubKey, cnr, req.Object)
		srv := New(noCallObjSvc, node)

		node.serverOutsideCnr = true
		node.clientOutsideCnr = true

		resp, err := srv.Replicate(context.Background(), req)
		require.NoError(t, err)
		require.EqualValues(t, 2048, resp.GetStatus().GetCode())
		require.Equal(t, "server does not match the object's storage policy", resp.GetStatus().GetMessage())

		node.serverOutsideCnr = false

		resp, err = srv.Replicate(context.Background(), req)
		require.NoError(t, err)
		require.EqualValues(t, 2048, resp.GetStatus().GetCode())
		require.Equal(t, "client does not match the object's storage policy", resp.GetStatus().GetMessage())
	})

	t.Run("local storage failure", func(t *testing.T) {
		node := newTestNode(t, serverPubKey, clientPubKey, cnr, req.Object)
		srv := New(noCallObjSvc, node)

		node.storeErr = errors.New("any error")

		resp, err := srv.Replicate(context.Background(), req)
		require.NoError(t, err)
		require.EqualValues(t, 1024, resp.GetStatus().GetCode())
		require.Equal(t, "failed to verify and store object locally: any error", resp.GetStatus().GetMessage())
	})

	t.Run("OK", func(t *testing.T) {
		node := newTestNode(t, serverPubKey, clientPubKey, cnr, req.Object)
		srv := New(noCallObjSvc, node)

		resp, err := srv.Replicate(context.Background(), req)
		require.NoError(t, err)
		require.EqualValues(t, 0, resp.GetStatus().GetCode())
		require.Empty(t, resp.GetStatus().GetMessage())
	})
}

type nopNode struct{}

func (x nopNode) ForEachContainerNodePublicKeyInLastTwoEpochs(cid.ID, func(pubKey []byte) bool) error {
	return nil
}

func (x nopNode) IsOwnPublicKey([]byte) bool {
	return false
}

func (x nopNode) VerifyAndStoreObject(object.Object) error {
	return nil
}

func BenchmarkServer_Replicate(b *testing.B) {
	ctx := context.Background()
	var node nopNode

	srv := New(nil, node)

	for _, tc := range []struct {
		name      string
		newSigner func(tb testing.TB) neofscrypto.Signer
	}{
		{
			name: "ECDSA SHA-512",
			newSigner: func(tb testing.TB) neofscrypto.Signer {
				return (*neofsecdsa.Signer)(randECDSAPrivateKey(tb))
			},
		},
		{
			name: "ECDSA SHA-256 deterministic",
			newSigner: func(tb testing.TB) neofscrypto.Signer {
				return (*neofsecdsa.SignerRFC6979)(randECDSAPrivateKey(tb))
			},
		},
		{
			name: "ECDSA SHA-256 WalletConnect",
			newSigner: func(tb testing.TB) neofscrypto.Signer {
				return (*neofsecdsa.SignerWalletConnect)(randECDSAPrivateKey(tb))
			},
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			req := anyValidRequest(b, tc.newSigner(b), cidtest.ID(), oidtest.ID())

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				resp, err := srv.Replicate(ctx, req)
				require.NoError(b, err)
				require.Zero(b, resp.GetStatus().GetCode())
			}
		})
	}
}
