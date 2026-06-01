package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())
	r.Use(prometheusMiddleware())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/api/v1")
	{
		v1.POST("/accounts", h.CreateAccount)
		v1.GET("/accounts/:id/summary", h.GetSummary)
		v1.POST("/accounts/:id/transactions", h.CreateTransaction)
	}

	return r
}
