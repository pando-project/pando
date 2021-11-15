package graphQL

import (
	"Pando/statetree/types"
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
)

var errUnexpectedType = "Unexpected type %T. expected %s"

var StateType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "State",
		Fields: graphql.Fields{
			"Cidlist": &graphql.Field{
				Type:    graphql.String,
				Resolve: State__Cidlist__resolve,
			},
			"LastUpdateHeight": &graphql.Field{
				Type:    graphql.Int,
				Resolve: State__LastHeight__resolve,
			},
			"LastUpdate": &graphql.Field{
				Type:    graphql.String,
				Resolve: State__LastUpdate__resolve,
			},
		},
	},
)

func State__Cidlist__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*types.ProviderStateRes)
	if ok {
		return ts.State.Cidlist, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "State.Cidlist")
	}
}
func State__LastHeight__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*types.ProviderStateRes)
	if ok {
		return ts.State.LastCommitHeight, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "State.LastUpdateHeight")
	}
}

func State__LastUpdate__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*types.ProviderStateRes)
	if ok {
		return ts.NewestUpdate, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "State.LastUpdate")
	}
}

var SnapShotType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "SnapShot",
		Fields: graphql.Fields{
			"Height": &graphql.Field{
				Type:    graphql.Int,
				Resolve: Ss__Height__resolve,
			},
			"CreateTime": &graphql.Field{
				Type:    graphql.String,
				Resolve: Ss__CreateTime__resolve,
			},
			"PreviousSnapShot": &graphql.Field{
				Type:    graphql.String,
				Resolve: Ss__PrevSs__resolve,
			},
			"ExtraInfo": &graphql.Field{
				Type:    graphql.String,
				Resolve: Ss__ExtraInfo__resolve,
			},
			"Update": &graphql.Field{
				Type:    graphql.String,
				Resolve: Ss__Update__resolve,
			},
		},
	})

func Ss__Height__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*types.SnapShot)
	if ok {
		return ts.Height, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "SnapShot.Height")
	}
}

func Ss__CreateTime__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*types.SnapShot)
	if ok {
		return ts.CreateTime, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "SnapShot.CreateTime")
	}
}

func Ss__PrevSs__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*types.SnapShot)
	if ok {
		return ts.PrevSnapShot, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "SnapShot.PrevSnapShot")
	}
}

func Ss__Update__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*types.SnapShot)
	if ok {
		str := ""
		for peer, update := range ts.Update {
			str += fmt.Sprintf("%s : %s", peer, update.String())
		}
		return str, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "SnapShot.Update")
	}
}
func Ss__ExtraInfo__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*types.SnapShot)
	jsonBytes, err := json.Marshal(ts.ExtraInfo)
	if err != nil {
		return nil, err
	}
	if ok {
		return fmt.Sprintf("%s", string(jsonBytes)), nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "SnapShot.ExtraInfo")
	}
}
