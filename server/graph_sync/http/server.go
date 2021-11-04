package http

import (
	"Pando/legs"
	"Pando/server/graphql"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"net"
	"net/http"
	"time"
)

var log = logging.Logger("graphsync")

type logEcapsulator struct {
	logger *logging.ZapEventLogger
}

type Server struct {
	server  *http.Server
	gserver *http.Server
	l       net.Listener
	gl      net.Listener
}

func New(listen string, glisten string, core *legs.LegsCore) (*Server, error) {
	// Create ingest HTTP server
	maddr, err := multiaddr.NewMultiaddr(listen)
	if err != nil {
		return nil, fmt.Errorf("bad ingest address in config %s: %s", listen, err)
	}
	graphSyncAddr, err := manet.ToNetAddr(maddr)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter().StrictSlash(true)
	server := &http.Server{
		Handler: r,
	}
	l, err := net.Listen("tcp", graphSyncAddr.String())
	if err != nil {
		return nil, err
	}
	s := &Server{server: server, l: l}

	h := newHandler(core)

	r.HandleFunc("/graph/sub/{peerid}", h.SubProvider)
	r.HandleFunc("/graph/get/{id}", h.GetGraph)

	if glisten != "" {
		maddr, err := multiaddr.NewMultiaddr(glisten)
		if err != nil {
			return nil, fmt.Errorf("bad ingest address in config %s: %s", listen, err)
		}
		graphQlAddr, err := manet.ToNetAddr(maddr)
		if err != nil {
			return nil, err
		}
		gl, err := net.Listen("tcp", graphQlAddr.String())
		if err != nil {
			return nil, err
		}
		s.gl = gl
		gqHandler, err := graphQL.GetHandler(core.DS, "")
		if err != nil {
			return nil, err
		}

		s.gserver = &http.Server{
			Handler:      gqHandler,
			WriteTimeout: 30 * time.Second,
			ReadTimeout:  30 * time.Second,
		}
	}

	return s, nil
}

func (s *Server) Start() error {
	log.Infow("graphsync http server listening", "listen_addr", s.l.Addr())
	if s.gserver != nil {
		go func() {
			log.Infow("graphql http server listening", "listen_addr", s.gl.Addr())
			err := s.gserver.Serve(s.gl)
			if err != nil {
				log.Errorf("graphql http server failed to start: %s", err.Error())
			}
		}()
	}
	err := s.server.Serve(s.l)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Info("graphsync http server shutdown")
	return s.server.Shutdown(ctx)
}
