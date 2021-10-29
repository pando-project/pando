package graphQL

import (
	task "Pando/mock_provider/task"
	logging "github.com/ipfs/go-log/v2"
	"strings"

	//"bytes"
	"context"
	"embed"
	"encoding/json"
	//"fmt"
	"github.com/ipfs/go-datastore"

	"github.com/graphql-go/graphql"
	"net/http"
)

//go:embed index.html
var index embed.FS

var log = logging.Logger("graphQl")

const nodeLoaderCtxKey = "NodeLoader"

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

func GetHandler(db datastore.Batching, accessToken string) (*http.ServeMux, error) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"Task": &graphql.Field{
					Name: "Task",
					Type: TaskType,
					Args: graphql.FieldConfigArgument{
						"UUID": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String), Description: "task uuid"},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						uuid := p.Args["UUID"].(string)
						tsk, err := db.Get(datastore.NewKey(uuid))
						if err != nil {
							return nil, err
						}
						t := new(task.FinishedTask)

						if err = json.Unmarshal(tsk, t); err != nil {
							_tsk := strings.Trim(string(tsk), "\"")
							_tsk = strings.ReplaceAll(_tsk, "\\", "")

							if err2 := json.Unmarshal([]byte(_tsk), t); err2 != nil {
								return nil, err
							} else {
								return t, nil
							}
						}
						return t, nil
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
