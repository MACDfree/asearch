package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type Config struct {
	IndexPath   string             `json:"indexPath"`
	Matches     []FileMatchPattern `json:"matches"`
	MaxFileSize int64              `json:"maxFileSizeMB"`
}

type FileMatchPattern struct {
	Paths    []string `json:"paths"`
	Patterns []string `json:"patterns"`
	Ignores  []string `json:"ignores"`
}

var Conf *Config = &Config{}

func init() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(errors.Wrap(err, "配置文件config.json打开失败"))
	}
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		panic(errors.Wrap(err, "配置文件config.json读取失败"))
	}
	err = json.Unmarshal(bs, Conf)
	if err != nil {
		panic(errors.Wrap(err, "配置文件config.json解析失败"))
	}
	Conf.IndexPath = filepath.Join(Conf.IndexPath, "asearch.index")
}
