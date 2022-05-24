package store

type Store interface {
	Put(bucketName string, key string, value any)
	Get(bucketName string, key string, value any)
	ForEachDelete(bucketName string, test func(key string, value []byte) bool)
}
