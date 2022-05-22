package filereader

import (
	"strings"

	"baliance.com/gooxml/document"
)

func wordRead(path string) (string, error) {
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
}

func init() {
	Regist("docx", FileReaderFunc(wordRead))
}
