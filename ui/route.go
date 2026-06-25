package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var distFS embed.FS

func Route(e *gin.Engine) {
	staticFS, _ := fs.Sub(distFS, "dist")
	httpFS := http.FS(staticFS)

	// 使用静态文件中间件处理所有 dist 目录下的文件
	// 这会自动处理 MIME 类型和缓存
	fileServer := http.StripPrefix("/", http.FileServer(httpFS))
	e.NoRoute(func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		// 检查文件是否存在于嵌入的文件系统中
		filePath := strings.TrimPrefix(path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		_, err := staticFS.Open(filePath)
		if err == nil {
			fileServer.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		// 如果文件不存在，则是 SPA 路由，提供 index.html
		ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		indexFile, err := staticFS.Open("index.html")
		if err != nil {
			ctx.String(http.StatusNotFound, "not found")
			return
		}
		defer indexFile.Close()
		stat, _ := indexFile.Stat()
		ctx.DataFromReader(http.StatusOK, stat.Size(), "text/html; charset=utf-8", indexFile, nil)
	})
}
