package http

import (
	"Pando/statetree"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"net"
	"net/http"
)

var log = logging.Logger("meta-server")

type Server struct {
	server *http.Server
	l      net.Listener
	gl     net.Listener
}

func New(listen string, stateTree *statetree.StateTree) (*Server, error) {
	// Create ingest HTTP server
	maddr, err := multiaddr.NewMultiaddr(listen)
	if err != nil {
		return nil, fmt.Errorf("bad ingest address in config %s: %s", listen, err)
	}
	metadataAddr, err := manet.ToNetAddr(maddr)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter().StrictSlash(true)
	server := &http.Server{
		Handler: r,
	}
	l, err := net.Listen("tcp", metadataAddr.String())
	if err != nil {
		return nil, err
	}
	s := &Server{server: server, l: l}

	h := newHandler(stateTree)

	r.HandleFunc("/meta/list", h.ListSnapShots)
	r.HandleFunc("/meta/info/{sscid}", h.ListSnapShotInfo)

	return s, nil
}

func (s *Server) Start() error {
	log.Infow("metadata http server listening", "listen_addr", s.l.Addr())
	err := s.server.Serve(s.l)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Info("metadata http server shutdown")
	return s.server.Shutdown(ctx)
}
