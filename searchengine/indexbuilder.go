package searchengine

import (
	"asearch/config"
	"asearch/filefinder"
	"asearch/filereader"
	"asearch/logger"
	"asearch/store/fileinfostore"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
)

func BuildIndex(indexPath string) {
	mapping := bleve.NewIndexMapping()

	err := mapping.AddCustomTokenizer("sego",
		map[string]interface{}{
			"dictpath": "dictionary.txt",
			"type":     "sego",
		},
	)
	if err != nil {
		panic(err)
	}
	err = mapping.AddCustomAnalyzer("sego",
		map[string]interface{}{
			"type":      "sego",
			"tokenizer": "sego",
		},
	)
	if err != nil {
		panic(err)
	}
	mapping.DefaultAnalyzer = "sego"

	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		fmt.Println(err)
		return
	}

	start := time.Now()
	fileInfos := filefinder.Find(config.Conf.Matches)
	var wg sync.WaitGroup
	for fileInfo := range fileInfos {
		fileinfostore.Put(fileInfo.Path,
			fileinfostore.FileMetaInfo{
				ModifiedTime: fileInfo.Document.ModifiedTime,
				UpdateTime:   time.Now(),
			})
		wg.Add(1)
		go func(f *filefinder.FileInfo) {
			start := time.Now()
			defer wg.Done()
			content, err := filereader.Read(f.Path)
			if err != nil {
				logger.Errorf("%+v\n", err)
			}
			f.Document.Content = content
			index.Index(f.Path, f.Document)
			fmt.Println(f.Path, time.Since(start))
		}(fileInfo)
	}
	wg.Wait()
	logger.Info("创建索引耗时：", time.Since(start))

	index.Close()
	runtime.GC()
}

func StartRebuildJob(index bleve.Index) {
	ticker := time.NewTicker(time.Duration(config.Conf.DelayHour) * time.Hour)
	go func(t *time.Ticker) {
		for {
			<-t.C
			logger.Info("开始进行索引更新")
			ReBuildIndex(index)
		}
	}(ticker)
}

func ReBuildIndex(index bleve.Index) {
	start := time.Now()
	fileInfos := filefinder.Find(config.Conf.Matches)
	var wg sync.WaitGroup
	for fileInfo := range fileInfos {
		modifiedTime := fileinfostore.Get(fileInfo.Path).ModifiedTime
		fileinfostore.Put(fileInfo.Path,
			fileinfostore.FileMetaInfo{
				ModifiedTime: fileInfo.Document.ModifiedTime,
				UpdateTime:   time.Now(),
			})
		if modifiedTime.Equal(fileInfo.Document.ModifiedTime) {
			continue
		}

		wg.Add(1)
		go func(f *filefinder.FileInfo) {
			start := time.Now()
			defer wg.Done()
			content, err := filereader.Read(f.Path)
			if err != nil {
				logger.Errorf("%+v\n", err)
			}
			f.Document.Content = content
			index.Index(f.Path, f.Document)
			fmt.Println(f.Path, time.Since(start))
		}(fileInfo)
	}
	wg.Wait()

	timeout := time.Since(start)
	fileinfostore.ForEach(func(path string, value fileinfostore.FileMetaInfo) bool {
		if time.Since(value.UpdateTime) > timeout+10*time.Second {
			index.Delete(path)
			logger.Info("删除索引", path)
			return true
		}
		return false
	})
	logger.Info("更新索引耗时：", time.Since(start))
}
