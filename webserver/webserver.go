package webserver

import (
	"asearch/logger"
	"asearch/store/fileinfostore"
	"asearch/util"
	"embed"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//go:embed template/*
var f embed.FS

func Run(addr string, index bleve.Index) {
	// 禁用终端打印颜色
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(Logger(), gin.Recovery())

	r.SetHTMLTemplate(getHTMLTemplate())

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "index.html", nil)
	})

	r.GET("/open", func(ctx *gin.Context) {
		path := ctx.Query("path")
		util.OpenLocal(path)
		ctx.JSON(200, gin.H{
			"message": "success",
		})
	})

	r.GET("/search", func(ctx *gin.Context) {
		str := ctx.Query("query")
		fromStr := ctx.Query("from")
		var from int
		if fromStr == "" {
			from = 0
		} else {
			var err error
			from, err = strconv.Atoi(fromStr)
			if err != nil {
				logger.Errorf("%+v", err)
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
		search.Size = 20
		// 按照修改时间倒序排
		search.SortBy([]string{"-ModifiedTime", "_score"})
		searchResults, err := index.Search(search)
		if err != nil {
			logger.Errorf("%+v", errors.Wrap(err, "查询失败"))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "查询失败",
			})
			return
		}

		list := make([]SearchResult, 0)
		for _, r := range searchResults.Hits {
			fileMetaInfo := fileinfostore.Get(r.ID)
			list = append(list, SearchResult{
				Path:         r.ID,
				ModifiedTime: fileMetaInfo.ModifiedTime,
				Fragments:    r.Fragments,
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
		if pages == 1 {
			pager = nil
		} else {
			for i := 0; i < int(pages); i++ {
				pager = append(pager, "search?query="+url.QueryEscape(str)+"&from="+strconv.Itoa(i*pageSize))
			}
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

	r.Run(addr)
}

func getHTMLTemplate() *template.Template {
	return template.Must(
		template.New("").Funcs(template.FuncMap{
			"unescaped": func(html string) any {
				return template.HTML(html)
			},
			"add": func(a, b int) any {
				return a + b
			},
			"timeformat": func(t time.Time) string {
				return t.Format("2006-01-02 15:04:05")
			},
		}).ParseFS(f, "template/*.html"))
}

type SearchResult struct {
	Path         string
	ModifiedTime time.Time
	Fragments    search.FieldFragmentMap
}
