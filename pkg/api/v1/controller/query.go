package controller

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func (c *Controller) Query(dbName string, queryStr string) (interface{}, error) {
	var bsonQuery bson.D
	err := bson.UnmarshalExtJSON([]byte(queryStr), true, &bsonQuery)
	if err != nil {
		return nil, err
	}

	var resJson bson.M
	opts := options.RunCmd().SetReadPreference(readpref.Primary())
	err = c.Core.StoreInstance.MetaStore.Database(dbName).
		RunCommand(context.TODO(), bsonQuery, opts).
		Decode(&resJson)
	if err != nil {
		return nil, err
	}
	logger.Debugf("data query return: %v", resJson)

	return resJson, err
}
