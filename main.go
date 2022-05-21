package main

import (
	"asearch/config"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
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

//go:embed template/*
var f embed.FS

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

	templ := template.Must(
		template.New("").Funcs(template.FuncMap{
			"unescaped": func(html string) any {
				return template.HTML(html)
			},
			"add": func(a, b int) any {
				return a + b
			},
		}).ParseFS(f, "template/*.html"))
	r.SetHTMLTemplate(templ)
	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "index.html", nil)
	})
	r.GET("/search", func(ctx *gin.Context) {
		str := ctx.Query("query")
		fromStr := ctx.Query("from")
		var from int
		if fromStr == "" {
			from = 0
		} else {
			from, err = strconv.Atoi(fromStr)
			if err != nil {
				log.Printf("%+v", err)
				from = 0
			}
		}
		if str == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "参数错误",
			})
			return
		}
		query := bleve.NewQueryStringQuery(str)
		search := bleve.NewSearchRequest(query)
		search.From = from
		search.Highlight = bleve.NewHighlight()
		search.Size = 10
		// 按照修改时间倒序排
		search.SortBy([]string{"-ModifiedTime", "_score"})
		searchResults, err := index.Search(search)
		if err != nil {
			log.Printf("%+v", errors.Wrap(err, "查询失败"))
			ctx.JSON(http.StatusInternalServerError, gin.H{
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
		total := searchResults.Total
		pageSize := searchResults.Request.Size
		currentPage := from/pageSize + 1
		pages := total / uint64(pageSize)
		if total%uint64(pageSize) != 0 {
			pages = pages + 1
		}
		pager := make([]string, 0, pages)
		for i := 0; i < int(pages); i++ {
			pager = append(pager, "search?query="+url.QueryEscape(str)+"&from="+strconv.Itoa(i*pageSize))
		}

		ctx.HTML(200, "search.html", gin.H{
			"list":        list,
			"total":       searchResults.Total,
			"took":        searchResults.Took,
			"pageSize":    searchResults.Request.Size,
			"currentPage": currentPage,
			"pager":       pager,
		})
	})
	r.GET("/open", func(ctx *gin.Context) {
		path := ctx.Query("path")
		openLocal(path)
		ctx.JSON(200, gin.H{
			"message": "success",
		})
	})
	go r.Run("127.0.0.1:9900")
	if config.Conf.OpenBrowserOnStart {
		openLocal("http://127.0.0.1:9900")
	}
	select {}
}

func openLocal(path string) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command(`cmd`, `/c`, `start`, path)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd.Start()
	}
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
			fileInfo.Document.Content = content
			fmt.Println(fileInfo.Path, "start")
			index.Index(fileInfo.Path, fileInfo.Document)
			fmt.Println(fileInfo.Path, "end")
		}
	}
	index.Close()
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
