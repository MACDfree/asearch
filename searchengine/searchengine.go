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

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/huichen/sego"
	"github.com/pkg/errors"
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

	for _, match := range config.Conf.Matches {
		for fileInfo := range filefinder.FindFiles(match.Paths, match.Patterns, match.Ignores) {
			fileinfostore.Put(fileInfo.Path, fileinfostore.FileMetaInfo{ModifiedTime: fileInfo.Document.ModifiedTime})
			content, err := filereader.Read(fileInfo.Path)
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
	runtime.GC()
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
