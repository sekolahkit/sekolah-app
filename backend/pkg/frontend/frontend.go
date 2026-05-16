package frontend

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var distFS embed.FS

func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/api/") {
			http.NotFound(w, r)
			return
		}

		if path != "/" && !strings.HasPrefix(path, "/assets/") {
			if _, err := fs.Stat(sub, strings.TrimPrefix(path, "/")); err != nil {
				r.URL.Path = "/"
			}
		}

		fileServer.ServeHTTP(w, r)
	})
}

func HasBuild() bool {
	_, err := fs.Stat(distFS, "dist/index.html")
	return err == nil
}
