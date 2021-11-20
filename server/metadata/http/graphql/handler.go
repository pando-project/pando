package graphQL

import (
	"Pando/statetree"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	//"bytes"
	"context"
	"embed"
	"encoding/json"

	"github.com/graphql-go/graphql"
	"net/http"
)

//go:embed index.html
var index embed.FS

var log = logging.Logger("graphQl")

type postData struct {
	Query     string                 `json:"query"`
	Operation string                 `json:"operation"`
	Variables map[string]interface{} `json:"variables"`
}

func CorsMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow cross domain AJAX requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		next(w, r)
	})
}

func GetHandler(st *statetree.StateTree) (*http.ServeMux, error) {
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
						pstate, err := st.GetProviderStateByPeerID(peerID)
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
						ss, err := st.GetSnapShot(sscid)
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

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(index)))
	mux.Handle("/graphql", CorsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var result *graphql.Result
		// todo
		//ctx := context.WithValue(r.Context(), nodeLoaderCtxKey, loader)
		ctx := context.Background()

		if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {
			var p postData
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			result = graphql.Do(graphql.Params{
				Context:        ctx,
				Schema:         schema,
				RequestString:  p.Query,
				VariableValues: p.Variables,
				OperationName:  p.Operation,
			})
		} else if r.Method == "POST" {
			err := r.ParseForm()
			if err != nil {
				log.Warnf("failed to read req: %v", err)
				return
			}
			result = graphql.Do(graphql.Params{
				Context:       ctx,
				Schema:        schema,
				RequestString: r.Form.Get("query"),
			})
		} else {
			result = graphql.Do(graphql.Params{
				Context:       ctx,
				Schema:        schema,
				RequestString: r.URL.Query().Get("query"),
			})
		}

		if len(result.Errors) > 0 {
			log.Infof("Query had errors: %s, %v", r.URL.Query().Get("query"), result.Errors)
		}
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Errorf("Failed to encode response: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	return mux, nil
}
