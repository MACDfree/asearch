package filereader

import (
	"strings"

	"baliance.com/gooxml/spreadsheet"
)

func excelRead(path string) (string, error) {
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
}

func init() {
	Regist("xlsx", FileReaderFunc(excelRead))
}
