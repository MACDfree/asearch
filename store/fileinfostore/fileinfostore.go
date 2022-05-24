package fileinfostore

import (
	"asearch/store"
	"time"
)

type FileMetaInfo struct {
	ModifiedTime time.Time
	UpdateTime   time.Time
}

func Put(path string, info FileMetaInfo) {
	store.GetStore().Put("fileinfo", path, info)
}

func Get(path string) FileMetaInfo {
	info := FileMetaInfo{}
	store.GetStore().Get("fileinfo", path, &info)
	return info
}

func ForEach(test func(path string, value FileMetaInfo) bool) {
	store.GetStore().ForEachDelete("fileinfo", func(key string, value []byte) bool {
		fmi := FileMetaInfo{}
		store.BytesTo(value, &fmi)
		return test(key, fmi)
	})
}
