package httpadminserver

import (
	"Pando/legs"
	"fmt"
	"github.com/gin-gonic/gin"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"net/http"
)

var log = logging.Logger("graphsync")

func New(listen string, core *legs.Core) (*http.Server, error) {
	// Create ingest HTTP server
	maddr, err := multiaddr.NewMultiaddr(listen)
	if err != nil {
		return nil, fmt.Errorf("bad ingest address in config %s: %s", listen, err)
	}
	adminAddr, err := manet.ToNetAddr(maddr)
	if err != nil {
		return nil, err
	}

	h := newHandler(core)

	r := gin.Default()
	legsRouter := r.Group("/admin/legs")
	legsRouter.GET("/subscribe", h.SubProvider)

	//ch := make(chan error)
	//go func() {
	//	err = r.Run(graphSyncAddr.String())
	//	ch <- err
	//}()

	return &http.Server{
		Addr:    adminAddr.String(),
		Handler: r,
	}, err
}
