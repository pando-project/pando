package p2phandler

import (
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando/pkg/api/core"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/api/v1/handler"
	"github.com/kenlabs/pando/pkg/api/v1/server/libp2p"
	pb "github.com/kenlabs/pando/pkg/api/v1/server/libp2p/proto"
	"github.com/kenlabs/pando/pkg/option"

	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"net/http"
)

var logger = logging.Logger("pando/libp2p")

// handler handles requests for the providers resource
type libp2pHandler struct {
	handler *handler.ServerHandler
}

// handlerFunc is the function signature required by handlers in this package
type handlerFunc func(context.Context, peer.ID, *pb.PandoMessage) ([]byte, error)

func NewHandler(core *core.Core, opt *option.Options) *libp2pHandler {
	return &libp2pHandler{
		handler: handler.New(core, opt),
	}
}

func (h *libp2pHandler) ProtocolID() protocol.ID {
	return libp2p.PandoProtocolID
}

func (h *libp2pHandler) HandleMessage(ctx context.Context, msgPeer peer.ID, msgbytes []byte) (proto.Message, error) {
	var req pb.PandoMessage
	err := proto.Unmarshal(msgbytes, &req)
	if err != nil {
		return nil, err
	}

	var handle handlerFunc
	var rspType pb.PandoMessage_MessageType
	switch req.GetType() {
	case pb.PandoMessage_GET_SNAPSHOT_CID_LIST:
		handle = h.metadataList
		rspType = pb.PandoMessage_GET_SNAPSHOT_CID_LIST_RESPONSE
	case pb.PandoMessage_GET_SNAPSHOT:
		handle = h.metadataSnapShot
		rspType = pb.PandoMessage_GET_SNAPSHOT_RESPONSE
	case pb.PandoMessage_REGISTER_PROVIDER:
		handle = h.providerRegister
		rspType = pb.PandoMessage_REGISTER_PROVIDER_RESPONSE
	case pb.PandoMessage_GET_PROVIDER_INFO:
		handle = h.listProviderInfo
		rspType = pb.PandoMessage_GET_PROVIDER_INFO_RESPONSE
	case pb.PandoMessage_GET_PROVIDER_HEAD:
		handle = h.listProviderHead
		rspType = pb.PandoMessage_GET_PROVIDER_HEAD_RESPONSE
	case pb.PandoMessage_GET_PANDO_INFO:
		handle = h.pandoInfo
		rspType = pb.PandoMessage_GET_PANDO_INFO_RESPONSE
	default:
		msg := "ussupported message type"
		logger.Errorw(msg, "type", req.GetType())
		return nil, fmt.Errorf("%s %d", msg, req.GetType())
	}

	data, err := handle(ctx, msgPeer, &req)
	if err != nil {
		err = HandleError(err, req.GetType().String())
		data = EncodeError(err)
		rspType = pb.PandoMessage_ERROR_RESPONSE
	}

	return &pb.PandoMessage{
		Type: rspType,
		Data: data,
	}, nil
}

func HandleError(err error, reqType string) *v1.Error {
	var apierr *v1.Error
	if errors.As(err, &apierr) {
		if apierr.Status() >= 500 {
			logger.Errorw(fmt.Sprint("cannot handle", reqType, "request"), "err", apierr.Error(), "status", apierr.Status())
			// Log the error and return only the 5xx status.
			return v1.NewError(nil, apierr.Status())
		}
	} else {
		apierr = v1.NewError(err, http.StatusBadRequest)
	}
	logger.Infow(fmt.Sprint("bad", reqType, "request"), "err", apierr.Error(), "status", apierr.Status())
	return apierr
}
