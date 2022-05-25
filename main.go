package main

import (
	"asearch/config"
	"asearch/logger"
	"asearch/notification"
	"asearch/searchengine"
	"asearch/util"
	"asearch/webserver"
	"os"

	"github.com/blevesearch/bleve/v2"
)

func main() {
	logger.Info("asearch开始执行")
	logger.Info("判断索引文件是否存在")
	_, err := os.Stat(config.Conf.IndexPath)
	if err != nil {
		logger.Infof("索引文件%s不存在", config.Conf.IndexPath)
		logger.Info("判断数据文件是否存在")
		_, err = os.Stat(config.Conf.DBPath)
		if err != nil {
			logger.Infof("数据文件%s不存在", config.Conf.DBPath)
		} else {
			logger.Infof("数据文件%s存在", config.Conf.DBPath)
			logger.Info("删除数据文件")
			os.RemoveAll(config.Conf.DBPath)
		}
		logger.Info("开始创建索引")
		searchengine.BuildIndex(config.Conf.IndexPath)
		logger.Info("结束创建索引，建议重新启动程序，以释放资源")
	} else {
		logger.Infof("索引文件%s存在", config.Conf.IndexPath)
		logger.Info("判断数据文件是否存在")
		_, err = os.Stat(config.Conf.DBPath)
		if err != nil {
			logger.Infof("数据文件%s不存在", config.Conf.DBPath)
			logger.Info("删除索引文件")
			os.RemoveAll(config.Conf.IndexPath)
			logger.Info("开始创建索引")
			searchengine.BuildIndex(config.Conf.IndexPath)
			logger.Info("结束创建索引，建议重新启动程序，以释放资源")
		} else {
			logger.Infof("数据文件%s存在", config.Conf.DBPath)
		}
	}

	logger.Info("打开索引文件")
	index, err := bleve.Open(config.Conf.IndexPath)
	if err != nil {
		panic(err)
	}
	logger.Info("执行一次索引更新操作")
	searchengine.ReBuildIndex(index)
	logger.Info("开始更新索引定时任务")
	searchengine.StartRebuildJob(index)

	logger.Info("准备提供搜索服务")
	go webserver.Run(config.Conf.Addr, index)
	if config.Conf.OpenBrowserOnStart {
		logger.Info("自动打开本地浏览器")
		util.OpenLocal("http://" + config.Conf.Addr)
	}
	notification.Run()
}
