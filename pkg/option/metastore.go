package option

import "go.mongodb.org/mongo-driver/mongo"

const (
	defaultMetaStoreType          = "mongodb"
	defaultMetaStoreConnectionURI = "mongodb://localhost:27018"
)

type MetaStore struct {
	Type          string `yaml:"Type"`
	ConnectionURI string `yaml:"ConnectionURI"`
	Client        *mongo.Client
}
