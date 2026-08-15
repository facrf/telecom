package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
)

//go:embed static/*
var files embed.FS

func Handler() http.Handler {
	static, _ := fs.Sub(files, "static")
	fileServer := http.FileServer(http.FS(static))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested := path.Clean(r.URL.Path)
		if requested == "." || requested == "/" {
			fileServer.ServeHTTP(w, r)
			return
		}
		if _, err := fs.Stat(static, requested[1:]); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// O FileServer redireciona /index.html para ./; servir a raiz permite que
		// ele encontre index.html sem criar um ciclo no fallback da SPA.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
