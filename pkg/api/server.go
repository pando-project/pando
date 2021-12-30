package api

import (
	"context"
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"net/http"
	"pando/pkg/api/core"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"pando/pkg/api/middleware"
	v1 "pando/pkg/api/v1"
	"pando/pkg/option"
)

var logger = logging.Logger("http-server")

type Server struct {
	Opt            *option.Options
	HttpServer     *http.Server
	HttpListenAddr string
	Core           *core.Core
}

func MustNewAPIServer(opt *option.Options, core *core.Core) *Server {
	router := gin.Default()
	//router.Use(middleware.ApiVersion())
	router.Use(middleware.LoggerToStdOut())
	v1API := v1.NewV1API(router, core)
	v1API.RegisterAPIs()

	httpMultiAddress, err := multiaddr.NewMultiaddr(opt.ServerAddress.HttpAPIListenAddress)
	if err != nil {
		panic("parse multiaddress failed")
	}
	httpListenAddress, err := manet.ToNetAddr(httpMultiAddress)
	if err != nil {
		return nil
	}
	return &Server{
		Opt: opt,
		HttpServer: &http.Server{
			Addr:    httpListenAddress.String(),
			Handler: router,
		},
		HttpListenAddr: httpListenAddress.String(),
		Core:           core,
	}
}

func (s *Server) Start() {
	logger.Infof("http server listening at: %s\n", s.HttpListenAddr)
	err := s.HttpServer.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("start http server failed: %v", err))
	}
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.HttpServer.Shutdown(ctx)
	if err != nil {
		panic(fmt.Sprintf("stop http server failed, forced to shutdown: %v", err))
	}

	logger.Infof("Bye! Pando!\n")
}
