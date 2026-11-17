package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"server_components/internal/app/auth"
	"server_components/internal/app/ds"
)

func (h *Handler) GetServerComponentsAPI(ctx *gin.Context) {
	currentUserID := auth.GetCurrentUserID()

	search := ctx.Query("search")
	if search == "" {
		search = ctx.Query("query")
	}

	minTDP := 0
	if minTDPStr := ctx.Query("min_tdp"); minTDPStr != "" {
		if val, err := strconv.Atoi(minTDPStr); err == nil && val >= 0 {
			minTDP = val
		}
	}

	components, err := h.Repository.GetPublishedComponentsAPI(search, minTDP, currentUserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(components),
		"data":   components,
	})
}

func (h *Handler) GetFeedComponentAPI(ctx *gin.Context) {
	currentUserID := auth.GetCurrentUserID()

	id := 0
	if idParam := ctx.Param("id"); idParam != "" {
		if val, err := strconv.Atoi(idParam); err == nil && val >= 0 {
			id = val
		}
	} else if idQuery := ctx.Query("id"); idQuery != "" {
		if val, err := strconv.Atoi(idQuery); err == nil && val >= 0 {
			id = val
		}
	}

	next := ctx.Query("next") == "true"

	feedItem, err := h.Repository.GetFeedComponentAPI(id, next, currentUserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   feedItem,
	})
}

func (h *Handler) GetUserDraftAPI(ctx *gin.Context) {
	currentUserID := auth.GetCurrentUserID()

	draft, err := h.Repository.GetUserDraftAPI(currentUserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "активный черновик у текущего пользователя не найден",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   draft,
	})
}

func (h *Handler) CreateDraftAPI(ctx *gin.Context) {
	currentUserID := auth.GetCurrentUserID()

	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "ошибка чтения multipart формы: " + err.Error(),
		})
		return
	}

	name := ctx.Request.FormValue("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "поле 'name' обязательно для заполнения",
		})
		return
	}

	imgHeader, _ := ctx.FormFile("image")
	videoHeader, _ := ctx.FormFile("video")

	draft, err := h.Repository.CreateDraftWithMedia(currentUserID, name, imgHeader, videoHeader)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "черновик успешно создан с загрузкой файлов в MinIO",
		"data":    draft,
	})
}

func (h *Handler) PublishComponentAPI(ctx *gin.Context) {
	currentUserID := auth.GetCurrentUserID()

	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "некорректный ID компонента, он должен быть > 0",
		})
		return
	}

	var req ds.PublishReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "некорректный формат JSON тела запроса: " + err.Error(),
		})
		return
	}

	if req.Description == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "краткая информация (description) обязательна для публикации",
		})
		return
	}
	if req.ComponentType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "тип компонента (component_type) обязателен для публикации",
		})
		return
	}
	if req.TDP <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "тепловыделение (tdp) должно быть больше 0",
		})
		return
	}

	comp, err := h.Repository.PublishComponentAPI(uint(id), currentUserID, req.Description, req.ComponentType, req.TDP)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "услуга успешно опубликована",
		"data":    comp,
	})
}

func (h *Handler) DeleteComponentAPI(ctx *gin.Context) {
	currentUserID := auth.GetCurrentUserID()

	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "некорректный ID компонента, он должен быть > 0",
		})
		return
	}

	err = h.Repository.SoftDeleteComponentAPI(uint(id), currentUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "услуга успешно логически удалена (soft delete)",
	})
}

func (h *Handler) ToggleLikeAPI(ctx *gin.Context) {
	currentUserID := auth.GetCurrentUserID()

	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "некорректный ID компонента, он должен быть > 0",
		})
		return
	}

	var req ds.LikeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "некорректный формат JSON тела запроса, требуется поле 'like': 0 или 1",
		})
		return
	}

	isLiked, likesCount, err := h.Repository.ToggleLikeAPI(uint(id), currentUserID, req.Like)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	actionMsg := "лайк успешно поставлен"
	if !isLiked {
		actionMsg = "лайк успешно отменен"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     actionMsg,
		"is_liked":    isLiked,
		"likes_count": likesCount,
	})
}

func (h *Handler) GetComponentWithCreatorAPI(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "некорректный ID компонента, он должен быть > 0",
		})
		return
	}

	fullComp, err := h.Repository.GetComponentWithCreator(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "услуга не найдена или удалена",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   fullComp,
	})
}

func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	var req ds.RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "поля 'login' и 'password' обязательны",
		})
		return
	}

	user, err := h.Repository.RegisterUser(req.Login, req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	logrus.Infof("Зарегистрирован новый пользователь: ID %d, Login: %s", user.ID, user.Login)
	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "пользователь успешно зарегистрирован",
		"data":    user,
	})
}

func (h *Handler) LoginUserAPI(ctx *gin.Context) {
	var req ds.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "поля 'login' и 'password' обязательны",
		})
		return
	}

	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil || user.Password != req.Password {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status":  "fail",
			"message": "неверный логин или пароль",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "аутентификация успешна (заглушка для ЛР 4)",
		"token":   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.dummy_token_lab3",
		"user":    user,
	})
}

func (h *Handler) LogoutUserAPI(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "деавторизация выполнена успешно (заглушка для ЛР 4)",
	})
}

func (h *Handler) GetUserComponentsAPI(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "некорректный ID пользователя, он должен быть > 0",
		})
		return
	}

	userComponents, err := h.Repository.GetUserWithComponents(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   userComponents,
	})
}
