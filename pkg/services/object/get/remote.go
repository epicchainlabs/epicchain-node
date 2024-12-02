package getsvc

import (
	"errors"

	"github.com/epicchainlabs/epicchain-node/pkg/core/client"
	apistatus "github.com/epicchainlabs/epicchain-sdk-go/client/status"
	objectSDK "github.com/epicchainlabs/epicchain-sdk-go/object"
	"go.uber.org/zap"
)

func (exec *execCtx) processNode(info client.NodeInfo) bool {
	exec.log.Debug("processing node...")

	client, ok := exec.remoteClient(info)
	if !ok {
		return true
	}

	obj, err := client.getObject(exec, info)

	var errSplitInfo *objectSDK.SplitInfoError

	switch {
	default:
		exec.status = statusUndefined
		exec.err = apistatus.ErrObjectNotFound

		exec.log.Debug("remote call failed",
			zap.String("error", err.Error()),
		)
	case err == nil:
		exec.status = statusOK
		exec.err = nil

		// both object and err are nil only if the original
		// request was forwarded to another node and the object
		// has already been streamed to the requesting party,
		// or it is a GETRANGEHASH forwarded request whose
		// response is not an object
		if obj != nil {
			exec.collectedObject = obj
			exec.writeCollectedObject()
		}
	case errors.Is(err, apistatus.Error) && !errors.Is(err, apistatus.ErrObjectNotFound):
		exec.status = statusAPIResponse
		exec.err = err
	case errors.As(err, &errSplitInfo):
		exec.status = statusVIRTUAL
		mergeSplitInfo(exec.splitInfo(), errSplitInfo.SplitInfo())
		exec.err = objectSDK.NewSplitInfoError(exec.infoSplit)
	}

	return exec.status != statusUndefined
}
