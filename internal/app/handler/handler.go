package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"server_components/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	
	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/server-components/feed")
	})

	router.GET("/server-components", h.GetFeed)
	router.GET("/server-components/feed", h.GetFeed)
	router.GET("/server-components/feed/:id", h.GetFeed)
	router.GET("/server-components/cards", h.GetCards)
	router.GET("/server-components/add", h.GetAddPage)

	router.POST("/server-components/add", h.CreateDraft)

	router.POST("/server-components/publish/:id", h.PublishComponent)

	router.POST("/server-components/delete/:id", h.DeleteComponent)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, statusCode int, message string) {
	logrus.Warnf("HTTP Error %d: %s", statusCode, message)
	ctx.HTML(statusCode, "error.html", gin.H{
		"message": message,
	})
}
