package graphQL

import (
	"Pando/mock_provider/task"
	"fmt"
	"github.com/graphql-go/graphql"
)

var errUnexpectedType = "Unexpected type %T. expected %s"

var TaskType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Task",
		Fields: graphql.Fields{
			"Status": &graphql.Field{
				Type:    graphql.Int,
				Resolve: Task__Status__resolve,
			},
			"Miner": &graphql.Field{
				Type:    graphql.String,
				Resolve: Task__Miner__resolve,
			},
			"MaxPriceAttoFIL": &graphql.Field{
				Type:    graphql.String,
				Resolve: Task__MPA__resolve,
			},
			"Verified": &graphql.Field{
				Type:    graphql.Boolean,
				Resolve: Task__Verified_resolve,
			},
			"FastRetrieval": &graphql.Field{
				Type:    graphql.Boolean,
				Resolve: Task_FR_resolve,
			},
		},
	},
)

func Task__Status__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*task.FinishedTask)
	if ok {
		return ts.Status, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "Task.Status")
	}
}
func Task__Miner__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*task.FinishedTask)
	if ok {
		return ts.Miner, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "Task.Miner")
	}
}

func Task__MPA__resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*task.FinishedTask)
	if ok {
		return ts.MaxPriceAttoFIL, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "Task.MaxPriceAttoFIL")
	}
}

func Task__Verified_resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*task.FinishedTask)
	if ok {
		return ts.Verified, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "Task.Verified")
	}
}

func Task_FR_resolve(p graphql.ResolveParams) (interface{}, error) {
	ts, ok := p.Source.(*task.FinishedTask)
	if ok {
		return ts.FastRetrieval, nil
	} else {
		return nil, fmt.Errorf(errUnexpectedType, p.Source, "Task.FastRetrieval")
	}
}
