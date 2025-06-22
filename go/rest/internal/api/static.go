package api

import (
	"embed"
	"io/fs"
	"net/http"
)

func StaticSite(embedFs embed.FS) http.Handler {
	contentFS, _ := fs.Sub(embedFs, "static")
	return http.FileServer(http.FS(contentFS))
}
