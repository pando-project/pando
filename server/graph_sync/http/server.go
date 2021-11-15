package http

import (
	"Pando/legs"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"net"
	"net/http"
)

var log = logging.Logger("graphsync")

type Server struct {
	server *http.Server
	l      net.Listener
}

func New(listen string, core *legs.LegsCore) (*Server, error) {
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

	return s, nil
}

func (s *Server) Start() error {
	log.Infow("graphsync http server listening", "listen_addr", s.l.Addr())

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
