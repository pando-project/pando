package controller

import (
	"github.com/pando-project/pando/pkg/api/v1/model"
	"strings"
)

func (c *Controller) PandoInfo() (*model.PandoInfo, error) {
	ipReplacer := func(multiAddress string, replaceIP string) string {
		splitAddress := strings.Split(multiAddress, "/")
		splitAddress[2] = replaceIP
		return strings.Join(splitAddress, "/")
	}

	return &model.PandoInfo{
		PeerID: c.Core.LegsCore.Host.ID().String(),
		Addresses: model.APIAddresses{
			HttpAPI:      ipReplacer(c.Options.ServerAddress.HttpAPIListenAddress, c.Options.ServerAddress.ExternalIP),
			GraphQLAPI:   ipReplacer(c.Options.ServerAddress.GraphqlListenAddress, c.Options.ServerAddress.ExternalIP),
			GraphSyncAPI: ipReplacer(c.Options.ServerAddress.P2PAddress, c.Options.ServerAddress.ExternalIP),
		},
	}, nil
}
