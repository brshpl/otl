package v1

import (
	"net/http"

	"github.com/brshpl/otl/internal/usecase"
	"github.com/brshpl/otl/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter -.
func NewRouter(handler *gin.Engine, l logger.Interface, t usecase.OneTimeLink) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// K8s probe
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	{
		r := newOneTimeLinkRoutes(h, t, l)
		handler.GET("/:link", r.getWithParam)
	}
}
