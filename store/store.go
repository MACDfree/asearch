package store

import "fmt"

type Store interface {
	Put(bucketName string, key string, value any)
	Get(bucketName string, key string, value any)
	ForEachDelete(bucketName string, test func(key string, value []byte) bool)
}

type KeyNotExistError struct {
	Key string
}

func (e *KeyNotExistError) Error() string {
	return fmt.Sprintf("key %s is not exist", e.Key)
}
