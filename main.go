package main

import (
	"asearch/config"
	"asearch/searchengine"
	"asearch/util"
	"asearch/webserver"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/blevesearch/bleve/v2"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("%+v", err)
			select {}
		}
	}()
	log.Println("asearch开始执行")
	log.Println("判断索引文件是否存在")
	_, err := os.Stat(config.Conf.IndexPath)
	reBuild := ""
	if err == nil {
		log.Printf("索引文件%s已存在\n", config.Conf.IndexPath)
		if config.Conf.Rebuild {
			fmt.Println("是否需要重建索引？y/n（默认：n）")
			fmt.Scanln(&reBuild)
			reBuild = strings.ToLower(reBuild)
			if reBuild == "" {
				reBuild = "n"
			}
		} else {
			reBuild = "n"
		}
	} else {
		reBuild = "y"
		log.Printf("索引文件%s不存在\n", config.Conf.IndexPath)
	}

	if strings.ToLower(reBuild) == "y" {
		log.Println("开始创建索引")
		searchengine.BuildIndex(config.Conf.IndexPath)
		log.Println("结束创建索引")
		log.Println("建议重新启动程序，以释放资源")
	}

	log.Println("打开索引文件")
	index, err := bleve.Open(config.Conf.IndexPath)
	if err != nil {
		panic(err)
	}
	searchengine.StartRebuildJob(index)

	log.Println("准备提供搜索服务")
	go webserver.Run("127.0.0.1:9900", index)
	if config.Conf.OpenBrowserOnStart {
		util.OpenLocal("http://127.0.0.1:9900")
	}
	select {}
}
