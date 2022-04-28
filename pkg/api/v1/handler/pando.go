package handler

import (
	"github.com/kenlabs/pando/pkg/api/v1/model"
	"strings"
)

func (h *ServerHandler) PandoInfo() (*model.PandoInfo, error) {
	ipReplacer := func(multiAddress string, replaceIP string) string {
		splitAddress := strings.Split(multiAddress, "/")
		splitAddress[2] = replaceIP
		return strings.Join(splitAddress, "/")
	}

	return &model.PandoInfo{
		PeerID: h.Core.LegsCore.Host.ID().String(),
		Addresses: model.APIAddresses{
			HttpAPI:      ipReplacer(h.Options.ServerAddress.HttpAPIListenAddress, h.Options.ServerAddress.ExternalIP),
			GraphQLAPI:   ipReplacer(h.Options.ServerAddress.GraphqlListenAddress, h.Options.ServerAddress.ExternalIP),
			GraphSyncAPI: ipReplacer(h.Options.ServerAddress.P2PAddress, h.Options.ServerAddress.ExternalIP),
		},
	}, nil
}
