package api

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"pando/pkg/api/core"
	"pando/pkg/option"
)

var logger = logging.Logger("http-server")

type Server struct {
	Opt  *option.Options
	Core *core.Core

	HttpServer     *http.Server
	HttpListenAddr string

	GraphqlServer     *http.Server
	GraphqlListenAddr string
}

func MustNewAPIServer(opt *option.Options, core *core.Core) (*Server, error) {
	httpMultiAddress, err := multiaddr.NewMultiaddr(opt.ServerAddress.HttpAPIListenAddress)
	if err != nil {
		return nil, err
	}
	httpListenAddress, err := manet.ToNetAddr(httpMultiAddress)
	if err != nil {
		return nil, err
	}

	graphqlMultiAddress, err := multiaddr.NewMultiaddr(opt.ServerAddress.GraphqlListenAddress)
	if err != nil {
		return nil, err
	}
	graphqlListenAddress, err := manet.ToNetAddr(graphqlMultiAddress)
	if err != nil {
		return nil, err
	}

	return &Server{
		Opt: opt,
		HttpServer: &http.Server{
			Addr:    httpListenAddress.String(),
			Handler: NewHttpRouter(core, opt),
		},
		HttpListenAddr: httpListenAddress.String(),
		GraphqlServer: &http.Server{
			Addr:    graphqlListenAddress.String(),
			Handler: NewGraphqlRouter(core),
		},
		GraphqlListenAddr: graphqlListenAddress.String(),
		Core:              core,
	}, nil
}

func (s *Server) StartHttpServer() error {
	logger.Infof("http server listening at: %s", s.HttpListenAddr)
	err := s.HttpServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) StopHttpServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("stop http server...")
	err := s.HttpServer.Shutdown(ctx)
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

func (s *Server) StopGraphqlServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("stop graphql server...")
	err := s.GraphqlServer.Shutdown(ctx)
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
	g := errgroup.Group{}
	g.Go(func() error {
		return s.StopHttpServer()
	})
	g.Go(func() error {
		return s.StopGraphqlServer()
	})
	err := g.Wait()
	if err != nil {
		return err
	}

	fmt.Println("Bye, Pando!")

	return nil
}
