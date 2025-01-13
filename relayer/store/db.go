package store

type KVStore interface {
	Get(key string) (string, error)
	Set(key string, value string) error
}
