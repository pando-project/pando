package http

import (
	"context"
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
)

var log = logging.Logger("metrics")

type Server struct {
	server   *http.Server
	listener net.Listener
}

func New(listen string) (*Server, error) {
	// read multi-address from config
	multiAddr, err := multiaddr.NewMultiaddr(listen)
	if err != nil {
		return nil, fmt.Errorf("bad metrics address in config %s: %s", listen, err)
	}
	metricAddr, err := manet.ToNetAddr(multiAddr)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", metricAddr.String())
	if err != nil {
		return nil, err
	}

	server := &Server{&http.Server{Addr: metricAddr.String(), Handler: promhttp.Handler()}, listener}

	return server, nil
}

func (s *Server) Start() error {
	log.Infow("metrics http server listening", "listen_addr", s.server.Addr)
	err := s.server.Serve(s.listener)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Info("metrics http server shutdown")
	return s.server.Shutdown(ctx)
}
