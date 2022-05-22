package filefinder

import (
	"path/filepath"
	"testing"
)

// func Test_findFiles(t *testing.T) {
// 	type args struct {
// 		paths   []string
// 		matchs  []string
// 		ignores []string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want <-chan *FileInfo
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := findFiles(tt.args.paths, tt.args.matchs, tt.args.ignores); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("findFiles() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_abc(t *testing.T) {
	t.Error(filepath.Glob("D:\\tmp\\安全整改\\**\\*.docx"))
}
