package graphql

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"

	"pando/pkg/statetree/types"
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
			"Cidlist": &graphql.Field{
				Type:    graphql.String,
				Resolve: StateCidListResolver,
			},
			"LastUpdateHeight": &graphql.Field{
				Type:    graphql.Int,
				Resolve: StateLastHeightResolver,
			},
			"LastUpdate": &graphql.Field{
				Type:    graphql.String,
				Resolve: StateLastUpdateResolver,
			},
		},
	},
)

func StateCidListResolver(params graphql.ResolveParams) (interface{}, error) {
	ts, ok := params.Source.(*types.ProviderStateRes)
	if ok {
		return ts.State.Cidlist, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "State.Cidlist")
	}
}
func StateLastHeightResolver(params graphql.ResolveParams) (interface{}, error) {
	ts, ok := params.Source.(*types.ProviderStateRes)
	if ok {
		return ts.State.LastCommitHeight, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "State.LastUpdateHeight")
	}
}

func StateLastUpdateResolver(params graphql.ResolveParams) (interface{}, error) {
	ts, ok := params.Source.(*types.ProviderStateRes)
	if ok {
		return ts.NewestUpdate, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "State.LastUpdate")
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
			"PreviousSnapShot": &graphql.Field{
				Type:    graphql.String,
				Resolve: SnapshotPreviousSnapshotResolver,
			},
			"ExtraInfo": &graphql.Field{
				Type:    graphql.String,
				Resolve: SnapshotExtraInfoResolver,
			},
			"Update": &graphql.Field{
				Type:    graphql.String,
				Resolve: SnapshotUpdateResolver,
			},
		},
	})

func SnapshotHeightResolver(params graphql.ResolveParams) (interface{}, error) {
	snapshot, ok := params.Source.(*types.SnapShot)
	if ok {
		return snapshot.Height, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "SnapShot.Height")
	}
}

func SnapshotCreateTimeResolver(params graphql.ResolveParams) (interface{}, error) {
	snapshot, ok := params.Source.(*types.SnapShot)
	if ok {
		return snapshot.CreateTime, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "SnapShot.CreateTime")
	}
}

func SnapshotPreviousSnapshotResolver(params graphql.ResolveParams) (interface{}, error) {
	snapshot, ok := params.Source.(*types.SnapShot)
	if ok {
		return snapshot.PrevSnapShot, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "SnapShot.PrevSnapShot")
	}
}

func SnapshotUpdateResolver(params graphql.ResolveParams) (interface{}, error) {
	snapshot, ok := params.Source.(*types.SnapShot)
	if ok {
		updateInfo := strings.Builder{}
		for peer, update := range snapshot.Update {
			updateInfo.WriteString(fmt.Sprintf("%s : %s\n", peer, update.String()))
		}
		return updateInfo, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "SnapShot.Update")
	}
}
func SnapshotExtraInfoResolver(params graphql.ResolveParams) (interface{}, error) {
	snapshot, ok := params.Source.(*types.SnapShot)
	jsonBytes, err := json.Marshal(snapshot.ExtraInfo)
	if err != nil {
		return nil, err
	}
	if ok {
		return string(jsonBytes), nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, params.Source, "SnapShot.ExtraInfo")
	}
}
