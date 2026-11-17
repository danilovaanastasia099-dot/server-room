package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"server_components/internal/app/repository"
)

const (
	descMaxRunes = 120 // ~2 строки предпросмотра описания
	userID       = 1   // фиксированный пользователь (без авторизации)
)

// ServiceWithLikes — вспомогательная структура для передачи компонента с кол-вом лайков в шаблон
type ServerComponentWithLikes struct {
	Service repository.ServerComponent
	Likes   int
	Liked   bool
}

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// GetFeed отображает ленту (один компонент в формате reels)
func (h *Handler) GetFeed(ctx *gin.Context) {
	idStr := ctx.Param("id")

	var serverComponent repository.ServerComponent
	var err error

	if idStr == "" {
		// Показываем первый опубликованный компонент
		components, e := h.Repository.GetPublishedServerComponents()
		if e != nil || len(components) == 0 {
			ctx.HTML(http.StatusOK, "error.html", gin.H{"message": "Нет доступных компонентов"})
			return
		}
		serverComponent = components[0]
	} else {
		id, e := strconv.Atoi(idStr)
		if e != nil {
			logrus.Error(e)
			ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"message": "Неверный ID"})
			return
		}

		// Если запрошен следующий — ищем следующий
		if ctx.Query("next") == "true" {
			serverComponent, err = h.Repository.GetNextServerComponent(id)
		} else {
			serverComponent, err = h.Repository.GetServerComponentByID(id)
		}
		if err != nil {
			logrus.Error(err)
			// При отсутствии следующего — показываем первый
			components, e2 := h.Repository.GetPublishedServerComponents()
			if e2 != nil || len(components) == 0 {
				ctx.HTML(http.StatusOK, "error.html", gin.H{"message": "Нет доступных компонентов"})
				return
			}
			serverComponent = components[0]
		}
	}

	// Обрезаем описание до ~2 строк
	desc := []rune(serverComponent.Description)
	shortDesc := serverComponent.Description
	expanded := ctx.Query("expand") == "true"
	if len(desc) > descMaxRunes {
		shortDesc = string(desc[:descMaxRunes]) + "..."
	}

	// URL для "Ещё" / "Меньше"
	moreURL := "/server-components/feed/" + strconv.Itoa(serverComponent.ID) + "?expand=true"
	lessURL := "/server-components/feed/" + strconv.Itoa(serverComponent.ID)

	// Проверяем, лайкнул ли текущий пользователь
	liked := false
	for _, uid := range serverComponent.Likes {
		if uid == userID {
			liked = true
			break
		}
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"serverComponent": serverComponent,
		"shortDesc":       shortDesc,
		"expanded":        expanded,
		"moreURL":         moreURL,
		"lessURL":         lessURL,
		"likes":           len(serverComponent.Likes),
		"liked":           liked,
	})
}

// GetAddPage отображает страницу добавления компонента
func (h *Handler) GetAddPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"time": time.Now().Format("15:04:05"),
	})
}

// AddServerComponent обрабатывает форму добавления компонента
func (h *Handler) AddServerComponent(ctx *gin.Context) {
	name := strings.TrimSpace(ctx.PostForm("name"))
	description := strings.TrimSpace(ctx.PostForm("description"))
	componentType := strings.TrimSpace(ctx.PostForm("type"))
	tdpStr := strings.TrimSpace(ctx.PostForm("tdp"))

	tdp, err := strconv.Atoi(tdpStr)
	if err != nil {
		tdp = 0
	}

	newComponent := repository.ServerComponent{
		Name:        name,
		Description: description,
		Type:        componentType,
		TDP:         tdp,
		ImageURL:    "",
		VideoURL:    "",
	}

	h.Repository.AddServerComponent(newComponent)

	ctx.Redirect(http.StatusFound, "/server-components")
}

// GetCards отображает плиточный каталог
func (h *Handler) GetCards(ctx *gin.Context) {
	minTDPStr := ctx.Query("min_tdp")

	var components []repository.ServerComponent
	var err error

	if minTDPStr != "" {
		minTDP, e := strconv.Atoi(minTDPStr)
		if e == nil {
			components, err = h.Repository.FilterServerComponentsByTDP(minTDP)
		} else {
			components, err = h.Repository.GetPublishedServerComponents()
		}
	} else {
		components, err = h.Repository.GetPublishedServerComponents()
	}

	if err != nil {
		logrus.Error(err)
		components = []repository.ServerComponent{}
	}

	// Преобразуем в ServerComponentWithLikes
	var serverComponentsWithLikes []ServerComponentWithLikes
	for _, c := range components {
		liked := false
		for _, uid := range c.Likes {
			if uid == userID {
				liked = true
				break
			}
		}
		serverComponentsWithLikes = append(serverComponentsWithLikes, ServerComponentWithLikes{
			Service: c,
			Likes:   len(c.Likes),
			Liked:   liked,
		})
	}

	ctx.HTML(http.StatusOK, "cards.html", gin.H{
		"serverComponents": serverComponentsWithLikes,
		"query":            minTDPStr,
	})
}

// ToggleLike переключает лайк (POST, без JS — redirect обратно)
func (h *Handler) ToggleLike(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/server-components")
		return
	}

	_, err = h.Repository.ToggleLike(id, userID)
	if err != nil {
		logrus.Error(err)
	}

	// Редирект обратно на ленту с этим компонентом
	referer := ctx.GetHeader("Referer")
	if referer == "" {
		referer = "/server-components/feed/" + idStr
	}
	ctx.Redirect(http.StatusFound, referer)
}
