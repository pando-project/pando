package handlers

import (
	"Pando/internal/handler"
	"Pando/internal/httpserver"
	"Pando/internal/metrics"
	"Pando/internal/registry"
	"Pando/statetree"
	"Pando/statetree/types"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type providersHttpHandler struct {
	adminHandler *handler.AdminHandler
}

func NewProvidersHandler(registry *registry.Registry) *providersHttpHandler {
	return &providersHttpHandler{
		adminHandler: handler.NewAdminHandler(registry),
	}
}

// RegisterProvider is the handlers of API: POST /providers
func (h *providersHttpHandler) RegisterProvider(c *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.RegisterProviderLatency)
	defer record()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorw("failed reading body", "err", err)

		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.adminHandler.RegisterProvider(body)
	if err != nil {
		httpserver.HandleError(c.Writer, err, "register")
		log.Warnf("register failed: %s", err.Error())
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}

// metaHttpHandler handles requests for the finder resource
type metaHttpHandler struct {
	stateTree   *statetree.StateTree
	graphSchema graphql.Schema
}

func NewMetaHandler(stateTree *statetree.StateTree) (*metaHttpHandler, error) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"State": &graphql.Field{
					Name: "State",
					Type: StateType,
					Args: graphql.FieldConfigArgument{
						"PeerID": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "account id of a provider"},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						peerIDStr := p.Args["PeerID"].(string)
						peerID, err := peer.Decode(peerIDStr)
						if err != nil {
							return nil, err
						}
						pstate, err := stateTree.GetProviderStateByPeerID(peerID)
						if err != nil {
							return nil, err
						}

						return pstate, nil
					},
				},
				"SnapShot": &graphql.Field{
					Name: "SnapShot",
					Type: SnapShotType,
					Args: graphql.FieldConfigArgument{
						"cid": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "cid of the snapshot"},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						cidStr := p.Args["cid"].(string)
						sscid, err := cid.Decode(cidStr)
						if err != nil {
							return nil, err
						}
						ss, err := stateTree.GetSnapShot(sscid)
						if err != nil {
							return nil, err
						}
						return ss, nil
					},
				},
			},
		},
		),
	})
	if err != nil {
		return nil, err
	}

	return &metaHttpHandler{
		stateTree,
		schema,
	}, nil
}

func (h *metaHttpHandler) ListSnapShotsList(c *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.ListMetadataLatency)
	defer record()

	snapCidList, err := h.stateTree.GetSnapShotCidList()
	if err != nil {
		log.Error("cannot list snapshots, err", err)
		http.Error(c.Writer, "", http.StatusInternalServerError)
		return
	}
	resBytes, err := json.Marshal(snapCidList)
	if err != nil {
		log.Error("cannot list snapshots, err", err)
		http.Error(c.Writer, "", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(c.Writer, http.StatusOK, resBytes)
}

func (h *metaHttpHandler) GetSnapShotInfo(c *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.ListSnapshotInfoLatency)
	defer record()

	heightStr := c.Query("height")
	cidStr := c.Query("sscid")

	if heightStr == "" && cidStr == "" {
		log.Error("nil snapshot cid and height")
		http.Error(c.Writer, "", http.StatusBadRequest)
		return
	}
	var ssCid cid.Cid
	var ssHeight uint64
	var err error
	var resCid *types.SnapShot
	var resHeight *types.SnapShot
	var resBytes []byte
	// get snapshot by cid
	if cidStr != "" {
		ssCid, err = cid.Decode(cidStr)
		if err != nil {
			log.Error("cannot decode input cid, err", err)
			http.Error(c.Writer, "", http.StatusBadRequest)
			return
		}
		resCid, err = h.stateTree.GetSnapShot(ssCid)
		if err != nil && err != statetree.NotFoundErr {
			log.Errorf("cannot get snapshot: %s, err: %s", ssCid.String(), err.Error())
			http.Error(c.Writer, "", http.StatusInternalServerError)
			return
		} else if err != nil {
			log.Errorf("not found snapshot: %s", ssCid.String())
			http.Error(c.Writer, "", http.StatusNotFound)
			return
		}
	}
	// get snapshot by height
	if heightStr != "" {
		ssHeight, err = strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
			log.Errorf("valid snapshout height: %s, err: %s", ssHeight, err)
			http.Error(c.Writer, "", http.StatusBadRequest)
			return
		}
		resHeight, err = h.stateTree.GetSnapShotByHeight(ssHeight)
		if err != nil && err != statetree.NotFoundErr {
			log.Warnf("cannot get snapshot by height: %s, err: %s", ssHeight, err.Error())
			http.Error(c.Writer, "", http.StatusInternalServerError)
			return
		} else if err != nil {
			log.Warnf("not found snapshot by height: %s", ssHeight)
			http.Error(c.Writer, "", http.StatusNotFound)
			return
		}
	}

	var finalRes *types.SnapShot
	// sure two snapshots are equal
	if heightStr != "" && cidStr != "" {
		if resCid == nil || resHeight == nil || resCid.CreateTime != resHeight.CreateTime {
			log.Warnf("snapshout: %s not matched height :%d", ssCid.String(), ssHeight)
			http.Error(c.Writer, "get different snapshots by cid and height", http.StatusBadRequest)
			return
		}
		finalRes = resCid
	} else {
		if resCid != nil {
			finalRes = resCid
		} else {
			finalRes = resHeight
		}
	}

	resBytes, err = json.Marshal(finalRes)
	if err != nil {
		log.Error("cannot marshal snapshot, err", err)
		http.Error(c.Writer, "", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(c.Writer, http.StatusOK, resBytes)
}

func (h *metaHttpHandler) GetPandoInfo(c *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.ListPandoInfoLatency)
	defer record()

	info, err := h.stateTree.GetPandoInfo()
	if err != nil {
		log.Error("cannot get pando info, err", err)
		resBytes, _ := json.Marshal(&httpserver.ErrorJsonResponse{
			Error: "failed to locate server status information",
		})
		WriteJsonResponse(c.Writer, http.StatusInternalServerError, resBytes)
		return
	}

	mas := strings.Fields(info.MultiAddrs)

	resBytes, err := json.Marshal(struct {
		PeerID     string
		Multiaddrs []string
	}{info.PeerID, mas})

	if err != nil {
		log.Error("cannot marshal pando info, err", err)
		http.Error(c.Writer, "", http.StatusInternalServerError)
		return
	}
	WriteJsonResponse(c.Writer, http.StatusOK, resBytes)
}

func WriteJsonResponse(w http.ResponseWriter, status int, body []byte) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if _, err := w.Write(body); err != nil {
		log.Errorw("cannot write response", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

var log = logging.Logger("public-server")

func (h *metaHttpHandler) HandleGraphql(c *gin.Context) {

	// allow cross domain AJAX requests
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	{
		var result *graphql.Result
		// todo
		//ctx := context.WithValue(r.Context(), nodeLoaderCtxKey, loader)
		ctx := context.Background()

		if c.Request.Header.Get("Content-Type") == "application/json" {
			var p postData
			defer c.Request.Body.Close()
			if err := json.NewDecoder(c.Request.Body).Decode(&p); err != nil {
				http.Error(c.Writer, err.Error(), http.StatusBadRequest)
				return
			}

			result = graphql.Do(graphql.Params{
				Context:        ctx,
				Schema:         h.graphSchema,
				RequestString:  p.Query,
				VariableValues: p.Variables,
				OperationName:  p.Operation,
			})
		} else {
			err := c.Request.ParseForm()
			if err != nil {
				log.Warnf("failed to read req: %v", err)
				return
			}
			result = graphql.Do(graphql.Params{
				Context:       ctx,
				Schema:        h.graphSchema,
				RequestString: c.Request.Form.Get("query"),
			})
		}

		if len(result.Errors) > 0 {
			log.Infof("Query had errors: %s, %v", c.Request.URL.Query().Get("query"), result.Errors)
		}
		if err := json.NewEncoder(c.Writer).Encode(result); err != nil {
			log.Errorf("Failed to encode response: %s", err)
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
