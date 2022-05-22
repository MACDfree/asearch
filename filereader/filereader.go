package filereader

import (
	"path/filepath"

	"github.com/pkg/errors"
)

func Read(path string) (string, error) {
	suffix := filepath.Ext(path)
	if suffix == "" {
		return "", errors.Errorf("无文件后缀，path=%s", path)
	}
	if f, ok := readerMap[suffix[1:]]; !ok {
		return "", errors.Errorf("无匹配的读取器，path=%s，ext=%s", path, suffix)
	} else {
		return f.Read(path)
	}
}

type FileReader interface {
	Read(path string) (string, error)
}

type FileReaderFunc func(string) (string, error)

func (r FileReaderFunc) Read(path string) (string, error) {
	return r(path)
}

func Regist(suffix string, reader FileReader) {
	readerMap[suffix] = reader
}

var readerMap = make(map[string]FileReader)
