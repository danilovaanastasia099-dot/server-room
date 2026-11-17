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

	api := router.Group("/api")
	{
		api.GET("/server-components", h.GetServerComponentsAPI)
		api.GET("/server-components/feed", h.GetFeedComponentAPI)
		api.GET("/server-components/feed/:id", h.GetFeedComponentAPI)
		api.GET("/server-components/draft", h.GetUserDraftAPI)
		api.GET("/server-components/:id", h.GetComponentWithCreatorAPI)
		api.POST("/server-components", h.CreateDraftAPI)
		api.PUT("/server-components/:id/publish", h.PublishComponentAPI)
		api.PUT("/server-components/:id", h.PublishComponentAPI)
		api.DELETE("/server-components/:id", h.DeleteComponentAPI)
		api.POST("/server-components/:id/like", h.ToggleLikeAPI)

		api.POST("/users/register", h.RegisterUserAPI)
		api.POST("/users/login", h.LoginUserAPI)
		api.POST("/users/logout", h.LogoutUserAPI)
		api.GET("/users/:id/components", h.GetUserComponentsAPI)
	}
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
