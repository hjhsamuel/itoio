package ui

import (
	"github.com/gin-gonic/gin"
)

func Route(e *gin.Engine) {
	e.StaticFile("/index.html", "ui/dist/index.html")
	e.StaticFile("/main.js", "ui/dist/main.js")
	e.StaticFile("/styles.css", "ui/dist/styles.css")
	e.StaticFile("/favicon.svg", "ui/dist/favicon.svg")
	e.StaticFile("/favicon.ico", "ui/dist/favicon.ico")
	e.StaticFile("/icons.svg", "ui/dist/icons.svg")
	e.Static("/assets", "ui/dist/assets")
	e.NoRoute(func(ctx *gin.Context) {
		ctx.File("ui/dist/index.html")
	})
}
