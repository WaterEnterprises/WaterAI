// Package embed provides embedded static assets for the Water AI server.
// This file uses Go's embed package to include the Next.js static export files
// in the final binary, enabling single-binary deployment.
package embed

import (
	"embed"
)

//go:embed all:out
var Frontend embed.FS
