package option

const defaultCacheStoreDir = "cachestore"

type CacheStore struct {
	Dir string `yaml:"Dir"`
}
