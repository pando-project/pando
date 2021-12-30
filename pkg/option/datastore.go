package option

const (
	defaultDataStoreType = "levelds"
	defaultDataStoreDir  = "datastore"
)

type DataStore struct {
	Type string `yaml:"Type"`
	Dir  string `yaml:"Dir"`
}
