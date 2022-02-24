package api

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	_ "net/http/pprof"
	"time"

	logging "github.com/ipfs/go-log/v2"

	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/option"
	"github.com/kenlabs/pando/pkg/util/multiaddress"
)

var logger = logging.Logger("http-server")

type Server struct {
	Opt  *option.Options
	Core *core.Core

	HttpServer     *http.Server
	HttpListenAddr string

	GraphqlServer     *http.Server
	GraphqlListenAddr string

	ProfileServer     *http.Server
	ProfileListenAddr string
}

func NewAPIServer(opt *option.Options, core *core.Core) (*Server, error) {
	httpListenAddress, err := multiaddress.MultiaddressToNetAddress(opt.ServerAddress.HttpAPIListenAddress)
	if err != nil {
		return nil, err
	}

	graphqlListenAddress, err := multiaddress.MultiaddressToNetAddress(opt.ServerAddress.GraphqlListenAddress)
	if err != nil {
		return nil, err
	}

	profileListenAddress, err := multiaddress.MultiaddressToNetAddress(opt.ServerAddress.ProfileListenAddress)
	if err != nil {
		return nil, err
	}

	return &Server{
		Opt:  opt,
		Core: core,

		HttpServer: &http.Server{
			Addr:    httpListenAddress,
			Handler: NewHttpRouter(core, opt),
		},
		HttpListenAddr: httpListenAddress,

		GraphqlServer: &http.Server{
			Addr:    graphqlListenAddress,
			Handler: NewGraphqlRouter(core),
		},
		GraphqlListenAddr: graphqlListenAddress,

		ProfileServer: &http.Server{
			Addr: profileListenAddress,
		},
		ProfileListenAddr: profileListenAddress,
	}, nil
}

func (s *Server) StartHttpServer() error {
	logger.Infof("http server listening at: %s", s.HttpListenAddr)
	return s.HttpServer.ListenAndServe()
}

func (s *Server) StopHttpServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("stop http server...")
	return s.HttpServer.Shutdown(ctx)
}

func (s *Server) StartGraphqlServer() error {
	logger.Infof("graphql server listening at: %s", s.GraphqlListenAddr)
	return s.GraphqlServer.ListenAndServe()
}

func (s *Server) StopGraphqlServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("stop graphql server...")
	return s.GraphqlServer.Shutdown(ctx)
}

func (s *Server) StartProfileServer() error {
	logger.Infof("profile server listening at: %s", s.ProfileListenAddr)
	return s.ProfileServer.ListenAndServe()
}

func (s *Server) StopProfileServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("stop profile server...")
	return s.ProfileServer.Shutdown(ctx)
}

func (s *Server) MustStartAllServers() {
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

	go func() {
		err := s.StartProfileServer()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("profile server cannot start: %v", err))
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
	g.Go(func() error {
		return s.StopProfileServer()
	})
	err := g.Wait()
	if err != nil {
		return err
	}

	fmt.Println("Bye, Pando!")

	return nil
}
