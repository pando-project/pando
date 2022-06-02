package option

const (
	defaultDataStoreType    = "levelds"
	defaultDataStoreDir     = "datastore"
	defaultSnapShotInterval = "60m"
)

type DataStore struct {
	Type             string `yaml:"Type"`
	Dir              string `yaml:"Dir"`
	SnapShotInterval string `yaml:"SnapShotInterval"`
}
