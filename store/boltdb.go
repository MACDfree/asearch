package store

import (
	"asearch/config"
	"bytes"
	"encoding/gob"
	"log"
	"sync"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

type BoltDB struct {
	db *bolt.DB
}

var instanceBoltDB *BoltDB
var onceBoltDB sync.Once

func GetStore() Store {
	once.Do(func() {
		db, err := bolt.Open(config.Conf.DBPath, 0600, nil)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		instanceBoltDB = &BoltDB{
			db: db,
		}
	})

	return instanceBoltDB
}

func (b *BoltDB) Put(bucketName string, key string, value any) {
	err := b.db.Update(func(tx *bolt.Tx) error {
		var b bytes.Buffer
		encoder := gob.NewEncoder(&b)
		if err := encoder.Encode(value); err != nil {
			return err
		}
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), b.Bytes())
	})
	if err != nil {
		panic(err)
	}
}

func (b *BoltDB) Get(bucketName string, key string, value any) {
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.Errorf("bucket %s 不存在", bucketName)
		}
		val := bucket.Get([]byte(key))
		if val == nil {
			return &KeyNotExistError{Key: key}
		}

		return BytesTo(val, value)
	})
	if err != nil {
		if _, ok := err.(*KeyNotExistError); ok {
			return
		}
		panic(err)
	}
}

func (b *BoltDB) ForEachDelete(bucketName string, test func(key string, value []byte) bool) {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.Errorf("bucket %s 不存在", bucketName)
		}

		deleteKeys := make([][]byte, 0, 10)
		bucket.ForEach(func(k, v []byte) error {
			if test(string(k), v) {
				deleteKeys = append(deleteKeys, k)
			}
			return nil
		})
		for _, deleteKey := range deleteKeys {
			bucket.Delete(deleteKey)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func BytesTo(value []byte, out any) error {
	d := gob.NewDecoder(bytes.NewReader(value))
	return d.Decode(out)
}
