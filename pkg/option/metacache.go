package option

import "go.mongodb.org/mongo-driver/mongo"

const (
	defaultMetaCacheType          = "mongodb"
	defaultMetaCacheConnectionURI = "mongodb://52.14.211.248:27018"
)

type MetaCache struct {
	Type          string `yaml:"Type"`
	ConnectionURI string `yaml:"ConnectionURI"`
	Client        *mongo.Client
}
