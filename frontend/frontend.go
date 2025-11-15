// Package frontend provides an HTTP handler to serve embedded frontend files.
package frontend

import (
	"embed"
	"net/http"
)

//go:embed css js index.html
var embeddedFiles embed.FS

// Handler returns an http.Handler that serves the embedded frontend files
func Handler() http.Handler {
	return http.FileServer(http.FS(embeddedFiles))
}
