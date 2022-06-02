package graphql

import (
	"fmt"
	"github.com/kenlabs/PandoStore/pkg/statestore/registry"
	"github.com/kenlabs/PandoStore/pkg/types/cbortypes"
	"strings"

	"github.com/graphql-go/graphql"
)

type postData struct {
	Query     string                 `json:"query"`
	Operation string                 `json:"operation"`
	Variables map[string]interface{} `json:"variables"`
}

var errUnexpectedType = "unexpected type %T. expected %s"

var StateType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "State",
		Fields: graphql.Fields{
			"PeerID": &graphql.Field{
				Type:    graphql.String,
				Resolve: StatePeerIDResolver,
			},
			"MetaList": &graphql.Field{
				Type:    graphql.String,
				Resolve: StateMetaListResolver,
			},
			"LastUpdateHeight": &graphql.Field{
				Type:    graphql.Int,
				Resolve: StateLastHeightResolver,
			},
		},
	},
)

func StatePeerIDResolver(params graphql.ResolveParams) (interface{}, error) {
	ts, ok := params.Source.(*registry.ProviderInfo)
	if ok {
		return ts.PeerID.String(), nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "ProviderInfo.PeerID")
	}
}

func StateMetaListResolver(params graphql.ResolveParams) (interface{}, error) {
	ts, ok := params.Source.(*registry.ProviderInfo)
	if ok {
		return ts.MetaList, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "ProviderInfo.MetaList")
	}
}
func StateLastHeightResolver(params graphql.ResolveParams) (interface{}, error) {
	ts, ok := params.Source.(*registry.ProviderInfo)
	if ok {
		return ts.LastUpdateHeight, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "ProviderInfo.LastUpdateHeight")
	}
}

var SnapShotType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "SnapShot",
		Fields: graphql.Fields{
			"Height": &graphql.Field{
				Type:    graphql.Int,
				Resolve: SnapshotHeightResolver,
			},
			"CreateTime": &graphql.Field{
				Type:    graphql.String,
				Resolve: SnapshotCreateTimeResolver,
			},
			"PrevSnapShot": &graphql.Field{
				Type:    graphql.String,
				Resolve: SnapshotPreviousSnapshotResolver,
			},
			"Update": &graphql.Field{
				Type:    graphql.String,
				Resolve: SnapshotUpdateResolver,
			},
		},
	})

func SnapshotHeightResolver(params graphql.ResolveParams) (interface{}, error) {
	snapshot, ok := params.Source.(*cbortypes.SnapShot)
	if ok {
		return snapshot.Height, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "SnapShot.Height")
	}
}

func SnapshotCreateTimeResolver(params graphql.ResolveParams) (interface{}, error) {
	snapshot, ok := params.Source.(*cbortypes.SnapShot)
	if ok {
		return snapshot.CreateTime, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "SnapShot.CreateTime")
	}
}

func SnapshotPreviousSnapshotResolver(params graphql.ResolveParams) (interface{}, error) {
	snapshot, ok := params.Source.(*cbortypes.SnapShot)
	if ok {
		return snapshot.PrevSnapShot, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "SnapShot.PrevSnapShot")
	}
}

func SnapshotUpdateResolver(params graphql.ResolveParams) (interface{}, error) {
	snapshot, ok := params.Source.(*cbortypes.SnapShot)
	if ok {
		updateInfo := strings.Builder{}
		for peer, update := range snapshot.Update {
			updateInfo.WriteString(fmt.Sprintf("{%s : %v},", peer, *update))
		}
		return updateInfo.String(), nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "SnapShot.Update")
	}
}
