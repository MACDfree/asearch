package filefinder

import (
	"asearch/config"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"
)

func FindFiles(paths []string, matchs []string, ignores []string) <-chan *FileInfo {
	filePaths := make(chan *FileInfo, 100)
	go func() {
		defer close(filePaths)
		for _, rootPath := range paths {
			err := filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
				if info.IsDir() {
					for _, ignore := range ignores {
						m, err := filepath.Match(ignore, info.Name())
						if err != nil {
							log.Printf("%+v\n", err)
							return nil
						}
						if m {
							return filepath.SkipDir
						}
					}
					return nil
				}
				if info.Size() > config.Conf.MaxFileSize*1024*1024 {
					return nil
				}
				// word临时文件夹忽略
				if strings.HasPrefix(info.Name(), "~$") {
					return nil
				}
				matched := false
				for _, match := range matchs {
					m, err := filepath.Match(match, info.Name())
					if err != nil {
						log.Printf("%+v\n", err)
						return nil
					}
					if m {
						matched = true
						break
					}
				}
				if !matched {
					return nil
				}
				for _, ignore := range ignores {
					m, err := filepath.Match(ignore, info.Name())
					if err != nil {
						log.Printf("%+v\n", err)
						return nil
					}
					if m {
						return nil
					}
				}
				filePaths <- &FileInfo{
					Path: path,
					Document: &FileDocument{
						Name:         info.Name(),
						ModifiedTime: info.ModTime(),
					},
				}
				return nil
			})
			if err != nil {
				log.Printf("%+v\n", err)
			}
		}
	}()
	return filePaths
}

type FileInfo struct {
	Path     string
	Document *FileDocument
}

type FileDocument struct {
	Name         string
	ModifiedTime time.Time
	Content      string
}
