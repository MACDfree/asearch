package store

import (
	"asearch/config"
	"bytes"
	"encoding/gob"
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
)

type BadgerDB struct {
	db *badger.DB
}

var instance *BadgerDB
var once sync.Once

func GetStore1() *BadgerDB {
	once.Do(func() {
		opt := badger.DefaultOptions(config.Conf.DBPath).
			WithCompression(options.Snappy).
			WithValueLogFileSize(10 * 1024 * 1024).
			WithNumLevelZeroTables(1).
			WithNumLevelZeroTablesStall(2).WithNumVersionsToKeep(0)
		db, err := badger.Open(opt)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		instance = &BadgerDB{
			db: db,
		}
		go instance.gc()
	})

	return instance
}

func (b *BadgerDB) Put(key string, value any) {
	err := b.db.Update(func(txn *badger.Txn) error {
		var b bytes.Buffer
		encoder := gob.NewEncoder(&b)
		if err := encoder.Encode(value); err != nil {
			return err
		}
		return txn.Set([]byte(key), b.Bytes())
	})
	if err != nil {
		panic(err)
	}
}

func (b *BadgerDB) Get(key string, value any) {
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		d := gob.NewDecoder(bytes.NewReader(val))
		return d.Decode(value)
	})
	if err != nil {
		panic(err)
	}
}

func (b *BadgerDB) gc() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
	again:
		err := b.db.RunValueLogGC(0.5)
		if err == nil {
			goto again
		}
	}
}
