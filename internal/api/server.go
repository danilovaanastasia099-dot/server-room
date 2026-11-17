package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"server_components/internal/app/handler"
	"server_components/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("Ошибка инициализации репозитория")
	}

	h := handler.NewHandler(repo)

	r := gin.Default()

	// Загрузка шаблонов
	r.LoadHTMLGlob("templates/*")

	// Раздача статики
	r.Static("/static", "./resources")

	// Маршруты
	r.GET("/server-components", h.GetFeed)
	r.GET("/server-components/feed", h.GetFeed)
	r.GET("/server-components/feed/:id", h.GetFeed)
	r.GET("/server-components/add", h.GetAddPage)
	r.POST("/server-components/add", h.AddServerComponent)
	r.GET("/server-components/cards", h.GetCards)
	r.POST("/server-components/like/:id", h.ToggleLike)

	// Редирект с корня на главную страницу
	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/server-components")
	})

	r.Run(":8080")
	log.Println("Server down")
}
