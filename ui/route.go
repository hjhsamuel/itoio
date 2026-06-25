package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
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
		// 增加安全响应头
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("X-Frame-Options", "DENY")
		ctx.Header("X-XSS-Protection", "1; mode=block")

		// 规范化路径并去除开头的斜杠
		requestPath := path.Clean(ctx.Request.URL.Path)
		filePath := strings.TrimPrefix(requestPath, "/")

		if filePath == "" || filePath == "." {
			filePath = "index.html"
		}

		// 禁止访问包含 .. 的路径（虽然 Clean 已经处理了，但为了安全再次检查）
		if strings.Contains(filePath, "..") {
			ctx.String(http.StatusBadRequest, "invalid path")
			return
		}

		// 检查文件是否存在且不是目录
		f, err := staticFS.Open(filePath)
		if err == nil {
			stat, err := f.Stat()
			f.Close()
			if err == nil && !stat.IsDir() {
				fileServer.ServeHTTP(ctx.Writer, ctx.Request)
				return
			}
		}

		// 如果文件不存在或者是目录，则是 SPA 路由，提供 index.html
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
