package fileinfostore

import (
	"asearch/store"
	"time"
)

type FileMetaInfo struct {
	ModifiedTime time.Time
}

func Put(path string, info FileMetaInfo) {
	store.GetStore().Put("fileinfo", path, info)
}

func Get(path string) FileMetaInfo {
	info := FileMetaInfo{}
	store.GetStore().Get("fileinfo", path, &info)
	return info
}
