package searchengine

import (
	"asearch/config"
	"asearch/filefinder"
	"asearch/filereader"
	"asearch/store/fileinfostore"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
)

func BuildIndex(indexPath string) {
	mapping := bleve.NewIndexMapping()
	os.RemoveAll(indexPath)

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
		fileinfostore.Put(fileInfo.Path, fileinfostore.FileMetaInfo{ModifiedTime: fileInfo.Document.ModifiedTime})
		wg.Add(1)
		go func(f *filefinder.FileInfo) {
			start := time.Now()
			defer wg.Done()
			content, err := filereader.Read(f.Path)
			if err != nil {
				log.Printf("%+v\n", err)
			}
			f.Document.Content = content
			index.Index(f.Path, f.Document)
			fmt.Println(f.Path, time.Since(start))
		}(fileInfo)
	}
	wg.Wait()
	log.Println("创建索引耗时：", time.Since(start))

	index.Close()
	runtime.GC()
}

func StartRebuildJob(index bleve.Index) {
	ticker := time.NewTicker(20 * time.Minute)
	go func(t *time.Ticker) {
		for {
			<-t.C
			log.Println("开始进行索引更新")
			reBuildIndex(index)
		}
	}(ticker)
}

func reBuildIndex(index bleve.Index) {
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
				log.Printf("%+v\n", err)
			}
			f.Document.Content = content
			index.Index(f.Path, f.Document)
			fmt.Println(f.Path, time.Since(start))
		}(fileInfo)
	}
	wg.Wait()

	fileinfostore.ForEach(func(path string, value fileinfostore.FileMetaInfo) bool {
		if time.Since(value.UpdateTime) > 40*time.Minute {
			index.Delete(path)
			log.Println("删除索引", path)
			return true
		}
		return false
	})
	log.Println("更新索引耗时：", time.Since(start))
}
