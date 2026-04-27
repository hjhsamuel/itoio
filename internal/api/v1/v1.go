package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/hjhsamuel/itoio/internal/core"
)

type api struct {
	s *core.Server
}

func Route(s *core.Server, g *gin.RouterGroup) {
	a := &api{s: s}

	gs := g.Group("/v1")
	ws(a, gs)
}
