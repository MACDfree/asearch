//go:build linux || darwin

package notification

import (
	_ "embed"
)

//go:embed icon.png
var icon []byte
