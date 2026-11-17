package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"server_components/internal/app/ds"
)

const (
	currentUserID = 1 
	descMaxRunes  = 120
)

func resolveMediaURLs(c *ds.ServerComponent) {
	
	if c.ImageURL == "" || strings.Contains(c.ImageURL, "invalid") {
		c.ImageURL = "/static/img/default.jpg"
	} else if strings.HasPrefix(c.ImageURL, "http://localhost:9000/server-images/") {
		fileName := strings.TrimPrefix(c.ImageURL, "http://localhost:9000/server-images/")
		if _, err := os.Stat(filepath.Join("resources", "img", fileName)); err == nil {
			c.ImageURL = "/static/img/" + fileName
		} else {
			c.ImageURL = "/static/img/default.jpg"
		}
	} else if strings.HasPrefix(c.ImageURL, "/static/img/") {
		fileName := strings.TrimPrefix(c.ImageURL, "/static/img/")
		if _, err := os.Stat(filepath.Join("resources", "img", fileName)); err != nil {
			c.ImageURL = "/static/img/default.jpg"
		}
	}

	if c.VideoURL == "" || strings.Contains(c.VideoURL, "invalid") {
		c.VideoURL = "/static/img/default.mp4"
	} else if strings.HasPrefix(c.VideoURL, "http://localhost:9000/server-videos/") {
		fileName := strings.TrimPrefix(c.VideoURL, "http://localhost:9000/server-videos/")
		if _, err := os.Stat(filepath.Join("resources", "img", fileName)); err == nil {
			c.VideoURL = "/static/img/" + fileName
		} else {
			c.VideoURL = "/static/img/default.mp4"
		}
	} else if strings.HasPrefix(c.VideoURL, "/static/img/") {
		fileName := strings.TrimPrefix(c.VideoURL, "/static/img/")
		if _, err := os.Stat(filepath.Join("resources", "img", fileName)); err != nil {
			c.VideoURL = "/static/img/default.mp4"
		}
	}
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	next := ctx.Query("next") == "true"

	var id int
	if idStr != "" {
		parsedID, err := strconv.Atoi(idStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, "Неверный идентификатор услуги")
			return
		}
		id = parsedID
	}

	comp, likesCount, err := h.Repository.GetFeedComponent(id, next)
	if err != nil {
		logrus.Warnf("GetFeedComponent error: %v", err)
		h.errorHandler(ctx, http.StatusNotFound, err.Error())
		return
	}

	resolveMediaURLs(comp)

	descRunes := []rune(comp.Description)
	shortDesc := comp.Description
	expanded := ctx.Query("expand") == "true"
	if len(descRunes) > descMaxRunes {
		shortDesc = string(descRunes[:descMaxRunes]) + "..."
	}

	moreURL := fmt.Sprintf("/server-components/feed/%d?expand=true", comp.ID)
	lessURL := fmt.Sprintf("/server-components/feed/%d", comp.ID)

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"serverComponent": comp,
		"shortDesc":       shortDesc,
		"expanded":        expanded,
		"moreURL":         moreURL,
		"lessURL":         lessURL,
		"likes":           likesCount,
	})
}

func (h *Handler) GetCards(ctx *gin.Context) {
	search := strings.TrimSpace(ctx.Query("query"))
	minTDPStr := strings.TrimSpace(ctx.Query("min_tdp"))

	minTDP := 0
	if minTDPStr != "" {
		if val, err := strconv.Atoi(minTDPStr); err == nil {
			minTDP = val
		}
	}

	cards, err := h.Repository.GetPublishedCards(search, minTDP)
	if err != nil {
		logrus.Errorf("GetPublishedCards error: %v", err)
		cards = nil
	}

	for i := range cards {
		resolveMediaURLs(&cards[i].Component)
	}

	ctx.HTML(http.StatusOK, "cards.html", gin.H{
		"serverComponents": cards,
		"query":            search,
		"min_tdp":          minTDPStr,
	})
}

func (h *Handler) GetAddPage(ctx *gin.Context) {
	draft, err := h.Repository.GetUserDraft(currentUserID)
	hasDraft := (err == nil && draft != nil)

	if hasDraft {
		resolveMediaURLs(draft)
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"hasDraft": hasDraft,
		"draft":    draft,
	})
}

func (h *Handler) CreateDraft(ctx *gin.Context) {
	name := strings.TrimSpace(ctx.PostForm("name"))
	if name == "" {
		h.errorHandler(ctx, http.StatusBadRequest, "Название компонента обязательно для заполнения")
		return
	}

	_, err := h.Repository.CreateDraft(currentUserID, name)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.Redirect(http.StatusFound, "/server-components/add")
}

func (h *Handler) PublishComponent(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Неверный ID компонента")
		return
	}

	description := strings.TrimSpace(ctx.PostForm("description"))
	componentType := strings.TrimSpace(ctx.PostForm("type"))
	tdpStr := strings.TrimSpace(ctx.PostForm("tdp"))

	tdp, _ := strconv.Atoi(tdpStr)

	err = h.Repository.PublishComponent(uint(id), currentUserID, description, componentType, tdp)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.Redirect(http.StatusFound, "/server-components/cards")
}

func (h *Handler) DeleteComponent(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Неверный ID для удаления")
		return
	}

	err = h.Repository.DeleteComponent(uint(id))
	if err != nil {
		logrus.Errorf("DeleteComponent error: %v", err)
	}

	redirectURL := ctx.Query("redirect")
	if redirectURL == "add" {
		ctx.Redirect(http.StatusFound, "/server-components/add")
		return
	}

	ctx.Redirect(http.StatusFound, "/server-components/cards")
}
