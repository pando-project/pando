package httppublicserver

import (
	"Pando/internal/metrics"
	"Pando/internal/registry"
	"Pando/server/handlers"
	"Pando/statetree"
	"fmt"
	coremetrics "github.com/filecoin-project/go-indexer-core/metrics"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"net/http"
)

var log = logging.Logger("admin-server")

func New(listen string, stateTree *statetree.StateTree, registry *registry.Registry) (*http.Server, error) {
	var err error

	// Create ingest HTTP server
	maddr, err := multiaddr.NewMultiaddr(listen)
	if err != nil {
		return nil, fmt.Errorf("bad ingest address in config %s: %s", listen, err)
	}
	publicAddr, err := manet.ToNetAddr(maddr)
	if err != nil {
		return nil, err
	}

	//l, err := net.Listen("tcp", adminAddr.String())
	if err != nil {
		return nil, err
	}
	providersHandler := handlers.NewProvidersHandler(registry)
	metaHandler, err := handlers.NewMetaHandler(stateTree)
	if err != nil {
		return nil, err
	}

	r := gin.Default()
	// Pando information
	r.GET("/status", metaHandler.GetPandoInfo)

	providerRouter := r.Group("/providers")
	providerRouter.POST("/register", providersHandler.RegisterProvider)

	// metadata(state-tree)
	metaRouter := r.Group("/meta")
	metaRouter.GET("/list", metaHandler.ListSnapShots)
	metaRouter.GET("/info", metaHandler.ListSnapShotInfo)

	// graphql(for state-tree)
	graphqlRouter := r.Group("/graphql")
	// todo get path auto
	r.LoadHTMLGlob("./server/_public/http/templates/*")
	graphqlRouter.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	graphqlRouter.POST("/search", metaHandler.HandleGraphql)

	// metrics
	metricsRouter := r.Group("/metrics")
	metricsRouter.GET("/", adapter.Wrap(func(h http.Handler) http.Handler {
		return metrics.Handler(coremetrics.DefaultViews)
	}))

	return &http.Server{
		Addr:    publicAddr.String(),
		Handler: r,
	}, err
}
