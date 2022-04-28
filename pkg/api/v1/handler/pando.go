package handler

import (
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/api/v1/model"
	"net/http"
	"strings"
)

func (h *ServerHandler) PandoInfo() (*model.PandoInfo, error) {

	pandoInfo, err := h.Core.StateTree.GetPandoInfo()
	if err != nil {
		return nil, v1.NewError(err, http.StatusNotFound)
	}

	ipReplacer := func(multiAddress string, replaceIP string) string {
		splitAddress := strings.Split(multiAddress, "/")
		splitAddress[2] = replaceIP
		return strings.Join(splitAddress, "/")
	}

	return &model.PandoInfo{
		PeerID: pandoInfo.PeerID,
		Addresses: model.APIAddresses{
			HttpAPI:      ipReplacer(h.Options.ServerAddress.HttpAPIListenAddress, h.Options.ServerAddress.ExternalIP),
			GraphQLAPI:   ipReplacer(h.Options.ServerAddress.GraphqlListenAddress, h.Options.ServerAddress.ExternalIP),
			GraphSyncAPI: ipReplacer(h.Options.ServerAddress.P2PAddress, h.Options.ServerAddress.ExternalIP),
		},
	}, nil
}
