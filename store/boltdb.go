package store

import (
	"asearch/config"
	"bytes"
	"encoding/gob"
	"log"
	"path/filepath"
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
		db, err := bolt.Open(filepath.Join(config.Conf.DBPath, "asearch.db"), 0600, nil)
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
			return errors.New("asearch bucket 不存在")
		}
		val := bucket.Get([]byte(key))
		if val == nil {
			return errors.Errorf("key %s 不存在", key)
		}

		d := gob.NewDecoder(bytes.NewReader(val))
		return d.Decode(value)
	})
	if err != nil {
		panic(err)
	}
}
