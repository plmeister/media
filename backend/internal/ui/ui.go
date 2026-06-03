// Package ui contains the embedded svelte map
package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed build/*
var EmbeddedUI embed.FS

func Handler() http.HandlerFunc {
	subFS, err := fs.Sub(EmbeddedUI, "build")
	if err != nil {
		panic(err)
	}

	index, _ := EmbeddedUI.ReadFile("build/index.html")

	fileServer := http.FileServer(http.FS(subFS))

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// serve static file if it exists
		f, err := subFS.Open(path)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(index)
	}
}
