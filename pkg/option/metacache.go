package option

import "go.mongodb.org/mongo-driver/mongo"

const (
	defaultMetaCacheType          = "mongodb"
	defaultMetaCacheConnectionURI = "mongodb://47.88.56.82:27018"
)

type MetaCache struct {
	Type          string `yaml:"Type"`
	ConnectionURI string `yaml:"ConnectionURI"`
	Client        *mongo.Client
}
