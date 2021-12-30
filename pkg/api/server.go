package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"golang.org/x/sync/errgroup"

	"pando/pkg/api/core"
	"pando/pkg/api/middleware"
	v1Graphql "pando/pkg/api/v1/graphql"
	v1Http "pando/pkg/api/v1/http"
	"pando/pkg/option"
)

var logger = logging.Logger("http-server")

type Server struct {
	Opt   *option.Options
	Core  *core.Core
	Group *errgroup.Group

	HttpServer     *http.Server
	HttpListenAddr string

	GraphqlServer     *http.Server
	GraphqlListenAddr string
}

func MustNewAPIServer(opt *option.Options, core *core.Core) *Server {
	httpRouter := gin.New()
	httpRouter.Use(middleware.WithLoggerFormatter())
	httpRouter.Use(middleware.WithCorsAllowAllOrigin())
	httpRouter.Use(gin.Recovery())

	v1HttpAPI := v1Http.NewV1HttpAPI(httpRouter, core)
	v1HttpAPI.RegisterAPIs()

	graphqlRouter := gin.New()
	graphqlRouter.Use(middleware.WithLoggerFormatter())
	graphqlRouter.Use(middleware.WithCorsAllowAllOrigin())
	graphqlRouter.Use(gin.Recovery())

	v1GraphAPI := v1Graphql.NewV1GraphqlAPI(graphqlRouter, core)
	v1GraphAPI.RegisterAPIs()

	httpMultiAddress, err := multiaddr.NewMultiaddr(opt.ServerAddress.HttpAPIListenAddress)
	if err != nil {
		panic("parse http multiaddress failed")
	}
	httpListenAddress, err := manet.ToNetAddr(httpMultiAddress)
	if err != nil {
		return nil
	}

	graphqlMultiAddress, err := multiaddr.NewMultiaddr(opt.ServerAddress.GraphqlListenAddress)
	if err != nil {
		panic("parse graphql multiaddress failed")
	}
	graphqlListenAddress, err := manet.ToNetAddr(graphqlMultiAddress)
	if err != nil {
		return nil
	}

	return &Server{
		Opt: opt,
		HttpServer: &http.Server{
			Addr:    httpListenAddress.String(),
			Handler: httpRouter,
		},
		HttpListenAddr: httpListenAddress.String(),
		GraphqlServer: &http.Server{
			Addr:    graphqlListenAddress.String(),
			Handler: graphqlRouter,
		},
		GraphqlListenAddr: graphqlListenAddress.String(),
		Core:              core,
		Group:             &errgroup.Group{},
	}
}

func (s *Server) StartHttpServer() error {
	logger.Infof("http server listening at: %s", s.HttpListenAddr)
	err := s.HttpServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) StartGraphqlServer() error {
	logger.Infof("graphql server listening at: %s", s.GraphqlListenAddr)
	err := s.GraphqlServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) StartAllServers() {
	go func() {
		err := s.StartHttpServer()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("http server cannot start: %v", err))
		}
	}()

	go func() {
		err := s.StartGraphqlServer()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("graphql server cannot start: %v", err))
		}
	}()
}

func (s *Server) StopAllServers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("stop http server...")
	err := s.HttpServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	fmt.Println("stop graphql server...")
	err = s.GraphqlServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Bye! Pando!")

	return nil
}
