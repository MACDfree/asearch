//go:build windows

package notification

import (
	_ "embed"
)

//go:embed iconwin.ico
var icon []byte
