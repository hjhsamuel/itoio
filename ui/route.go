package ui

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var distFS embed.FS

func Route(e *gin.Engine) {
	staticFS, _ := fs.Sub(distFS, "dist")
	httpFS := http.FS(staticFS)

	e.StaticFileFS("/index.html", "index.html", httpFS)
	e.StaticFileFS("/main.js", "main.js", httpFS)
	e.StaticFileFS("/styles.css", "styles.css", httpFS)
	e.StaticFileFS("/favicon.svg", "favicon.svg", httpFS)
	e.StaticFileFS("/favicon.ico", "favicon.ico", httpFS)
	e.StaticFileFS("/icons.svg", "icons.svg", httpFS)
	e.StaticFS("/assets", httpFS)

	e.NoRoute(func(ctx *gin.Context) {
		ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		f, err := staticFS.Open("index.html")
		if err != nil {
			ctx.String(http.StatusNotFound, "not found")
			return
		}
		defer f.Close()
		stat, _ := f.Stat()
		ctx.DataFromReader(http.StatusOK, stat.Size(), "text/html; charset=utf-8", f, nil)
	})
}
