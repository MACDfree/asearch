package main

import (
	"asearch/config"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"baliance.com/gooxml/document"
	"baliance.com/gooxml/spreadsheet"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/gin-gonic/gin"
	"github.com/huichen/sego"
	"github.com/pkg/errors"
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
		fmt.Println("是否需要重建索引？（默认：n）")
		fmt.Scanln(&reBuild)
		reBuild = strings.ToLower(reBuild)
		if reBuild == "" {
			reBuild = "n"
		}
	} else {
		reBuild = "y"
		log.Printf("索引文件%s不存在\n", config.Conf.IndexPath)
	}
	if reBuild == "y" {
		log.Println("开始创建索引")
		BuildIndex(config.Conf.IndexPath)
		log.Println("结束创建索引")
		log.Println("建议重新启动程序，以释放资源")
	}

	log.Println("打开索引文件")
	index, err := bleve.Open(config.Conf.IndexPath)
	if err != nil {
		panic(err)
	}

	log.Println("准备提供搜索服务")
	r := gin.Default()
	r.GET("/search", func(c *gin.Context) {
		str := c.Query("query")
		if str == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "参数错误",
			})
			return
		}
		query := bleve.NewQueryStringQuery(str)
		search := bleve.NewSearchRequest(query)
		search.Highlight = bleve.NewHighlight()
		search.Size = 10
		searchResults, err := index.Search(search)
		if err != nil {
			log.Printf("%+v", errors.Wrap(err, "查询失败"))
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "查询失败",
			})
			return
		}

		list := make([]SearchResult, 0)
		for _, r := range searchResults.Hits {
			list = append(list, SearchResult{
				Path:      r.ID,
				Fragments: r.Fragments,
			})
		}
		c.JSON(200, gin.H{
			"list":     list,
			"total":    searchResults.Total,
			"took":     searchResults.Took,
			"pageSize": searchResults.Request.Size,
		})
	})
	r.Run("127.0.0.1:9900")
}

type SearchResult struct {
	Path      string
	Fragments search.FieldFragmentMap
}

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

	for _, match := range config.Conf.Matches {
		for fileInfo := range findFiles(match.Paths, match.Patterns, match.Ignores) {
			content, err := readFile(fileInfo.Path)
			if err != nil {
				log.Printf("%+v\n", err)
			}
			fileInfo.Content = content
			fmt.Println(fileInfo.Path, "start")
			index.Index(fileInfo.Path, fileInfo)
			fmt.Println(fileInfo.Path, "end")
		}
	}
	index.Close()
}

type FileInfo struct {
	Path         string
	ModifiedTime time.Time
	Content      string
}

func findFiles(paths []string, matchs []string, ignores []string) <-chan *FileInfo {
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
					Path:         path,
					ModifiedTime: info.ModTime(),
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

func readFile(path string) (string, error) {
	if strings.HasSuffix(path, ".docx") {
		doc, err := document.Open(path)
		if err != nil {
			return "", err
		}
		var sb strings.Builder
		for _, p := range doc.Paragraphs() {
			for _, r := range p.Runs() {
				sb.WriteString(r.Text())
			}
			sb.WriteByte('\n')
		}
		return sb.String(), nil
	} else if strings.HasSuffix(path, ".xlsx") {
		xls, err := spreadsheet.Open(path)
		if err != nil {
			return "", nil
		}
		var sb strings.Builder
		for _, sheet := range xls.Sheets() {
			for _, row := range sheet.Rows() {
				for _, cell := range row.Cells() {
					sb.WriteString(cell.GetString())
					sb.WriteByte(' ')
				}
				sb.WriteByte('\n')
			}
			sb.WriteByte('\n')
		}
		return sb.String(), nil
	} else if strings.HasSuffix(path, ".txt") {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	return "", errors.New("无法解析文件内容")
}

type PathMatcher interface {
	Match(match []string) bool
}

func init() {
	registry.RegisterAnalyzer("sego", analyzerConstructor)
	registry.RegisterTokenizer("sego", tokenizerConstructor)
}

type SegoTokenizer struct {
	tker sego.Segmenter
}

func (s *SegoTokenizer) loadDictory(dict string) {
	s.tker.LoadDictionary(dict)
}

func (s *SegoTokenizer) Tokenize(sentence []byte) analysis.TokenStream {
	result := make(analysis.TokenStream, 0)
	// words := s.tker.InternalSegment(sentence, true)
	words := s.tker.Segment(sentence)
	for pos, word := range words {
		// fmt.Println(word.Token().Text())
		token := analysis.Token{
			Start:    word.Start(),
			End:      word.End(),
			Position: pos + 1,
			Term:     []byte(word.Token().Text()),
			Type:     analysis.Ideographic,
		}
		result = append(result, &token)
	}
	return result
}

func tokenizerConstructor(config map[string]interface{}, cache *registry.Cache) (analysis.Tokenizer, error) {
	dictpath, ok := config["dictpath"].(string)
	if !ok {
		return nil, errors.New("config dictpath not found")
	}
	tokenizer := &SegoTokenizer{}
	tokenizer.loadDictory(dictpath)
	return tokenizer, nil
}

type SegoAnalyzer struct{}

func analyzerConstructor(config map[string]interface{}, cache *registry.Cache) (*analysis.Analyzer, error) {
	tokenizerName, ok := config["tokenizer"].(string)
	if !ok {
		return nil, errors.New("must specify tokenizer")
	}
	tokenizer, err := cache.TokenizerNamed(tokenizerName)
	if err != nil {
		return nil, err
	}
	alz := &analysis.Analyzer{
		Tokenizer: tokenizer,
	}
	return alz, nil
}
