package api

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/hjhsamuel/itoio/internal/api/v1"
	"github.com/hjhsamuel/itoio/internal/core"
)

func Route(s *core.Server, e *gin.Engine) {
	g := e.Group("/ito")

	v1.Route(s, g)
}
