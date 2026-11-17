package repository

import (
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"gorm.io/gorm"
	"server_components/internal/app/ds"
)

type CardItemWithLikes struct {
	Component ds.ServerComponent
	Likes     int64
}

func (r *Repository) GetPublishedComponentsAPI(search string, minTDP int, currentUserID uint) ([]ds.ComponentListItem, error) {
	var components []ds.ServerComponent

	tx := r.db.Where("status = ?", "опубликован")

	if search != "" {
		tx = tx.Where("name ILIKE ?", "%"+search+"%")
	}
	if minTDP > 0 {
		tx = tx.Where("tdp >= ?", minTDP)
	}

	err := tx.Order("id asc").Find(&components).Error
	if err != nil {
		return nil, err
	}

	result := make([]ds.ComponentListItem, 0, len(components))
	for _, comp := range components {
		var likesCount int64
		_ = r.db.Model(&ds.ComponentLike{}).Where("component_id = ?", comp.ID).Count(&likesCount).Error

		isCreator := 0
		if comp.CreatorID == currentUserID {
			isCreator = 1
		}

		result = append(result, ds.ComponentListItem{
			ServerComponent: comp,
			IsCreator:       isCreator,
			LikesCount:      likesCount,
		})
	}

	return result, nil
}

func (r *Repository) GetFeedComponentAPI(id int, next bool, currentUserID uint) (*ds.ComponentFeedItem, error) {
	var comp ds.ServerComponent

	if id == 0 {
		err := r.db.Where("status = ?", "опубликован").Order("id asc").Limit(1).First(&comp).Error
		if err != nil {
			return nil, err
		}
	} else if next {
		err := r.db.Where("id > ? AND status = ?", id, "опубликован").Order("id asc").Limit(1).First(&comp).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = r.db.Where("status = ?", "опубликован").Order("id asc").Limit(1).First(&comp).Error
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	} else {
		err := r.db.Where("id = ?", id).Limit(1).First(&comp).Error
		if err != nil {
			return nil, err
		}
		if comp.Status == "удален" {
			return nil, fmt.Errorf("данная услуга удалена и недоступна для просмотра")
		}
		if comp.Status != "опубликован" {
			return nil, fmt.Errorf("услуга находится в статусе '%s'", comp.Status)
		}
	}

	var likesCount int64
	_ = r.db.Model(&ds.ComponentLike{}).Where("component_id = ?", comp.ID).Count(&likesCount).Error

	var userLikeCount int64
	_ = r.db.Model(&ds.ComponentLike{}).Where("component_id = ? AND user_id = ?", comp.ID, currentUserID).Count(&userLikeCount).Error
	isLiked := 0
	if userLikeCount > 0 {
		isLiked = 1
	}

	return &ds.ComponentFeedItem{
		ServerComponent: comp,
		IsLiked:         isLiked,
		LikesCount:      likesCount,
	}, nil
}

func (r *Repository) GetUserDraftAPI(userID uint) (*ds.ServerComponent, error) {
	var comp ds.ServerComponent
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").Limit(1).First(&comp).Error
	if err != nil {
		return nil, err
	}
	return &comp, nil
}

func (r *Repository) CreateDraftWithMedia(userID uint, name string, imgHeader, videoHeader *multipart.FileHeader) (*ds.ServerComponent, error) {
	var count int64
	r.db.Model(&ds.ServerComponent{}).Where("creator_id = ? AND status = ?", userID, "черновик").Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("у вас уже есть активный черновик. Завершите его публикацию перед созданием нового")
	}

	imageURL := "http://" + r.minioEndpoint + "/" + r.minioImageBucket + "/default.jpg"
	videoURL := "http://" + r.minioEndpoint + "/" + r.minioVideoBucket + "/default.mp4"

	if imgHeader != nil {
		uploadedImgURL, err := r.UploadFileToMinIO(r.minioImageBucket, "img", userID, imgHeader)
		if err == nil && uploadedImgURL != "" {
			imageURL = uploadedImgURL
		}
	}

	if videoHeader != nil {
		uploadedVidURL, err := r.UploadFileToMinIO(r.minioVideoBucket, "vid", userID, videoHeader)
		if err == nil && uploadedVidURL != "" {
			videoURL = uploadedVidURL
		}
	}

	draft := ds.ServerComponent{
		Name:          name,
		Description:   "",
		Status:        "черновик",
		ImageURL:      imageURL,
		VideoURL:      videoURL,
		ComponentType: "",
		TDP:           0,
		CreatedAt:     time.Now(),
		CreatorID:     userID,
		FormedAt:      nil,
	}

	err := r.db.Create(&draft).Error
	if err != nil {
		return nil, err
	}

	return &draft, nil
}

func (r *Repository) PublishComponentAPI(id uint, userID uint, description, componentType string, tdp int) (*ds.ServerComponent, error) {
	var comp ds.ServerComponent
	err := r.db.Where("id = ? AND creator_id = ? AND status = ?", id, userID, "черновик").First(&comp).Error
	if err != nil {
		return nil, fmt.Errorf("черновик не найден, принадлежит другому пользователю или уже опубликован")
	}

	now := time.Now()
	comp.Description = description
	comp.ComponentType = componentType
	comp.TDP = tdp
	comp.Status = "опубликован"
	comp.FormedAt = &now

	err = r.db.Save(&comp).Error
	if err != nil {
		return nil, err
	}

	return &comp, nil
}

func (r *Repository) SoftDeleteComponentAPI(id uint, userID uint) error {
	var comp ds.ServerComponent
	err := r.db.Where("id = ?", id).First(&comp).Error
	if err != nil {
		return fmt.Errorf("услуга с id %d не найдена: %w", id, err)
	}

	if comp.CreatorID != userID {
		return fmt.Errorf("доступ запрещен: вы можете удалять только созданные вами услуги")
	}

	if comp.Status == "удален" {
		return fmt.Errorf("услуга уже удалена")
	}

	comp.Status = "удален"
	return r.db.Save(&comp).Error
}

func (r *Repository) ToggleLikeAPI(componentID uint, userID uint, likeValue int) (bool, int64, error) {
	var comp ds.ServerComponent
	err := r.db.Where("id = ? AND status != ?", componentID, "удален").First(&comp).Error
	if err != nil {
		return false, 0, fmt.Errorf("компонент не найден или удален")
	}

	var existingLike ds.ComponentLike
	findErr := r.db.Where("user_id = ? AND component_id = ?", userID, componentID).First(&existingLike).Error

	if likeValue == 1 {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			newLike := ds.ComponentLike{
				UserID:      userID,
				ComponentID: componentID,
			}
			if err := r.db.Create(&newLike).Error; err != nil {
				return false, 0, err
			}
		}
	} else if likeValue == 0 {
		if findErr == nil {
			if err := r.db.Delete(&existingLike).Error; err != nil {
				return false, 0, err
			}
		}
	} else {
		return false, 0, fmt.Errorf("поле like должно быть 0 (отмена) или 1 (лайк)")
	}

	var likesCount int64
	_ = r.db.Model(&ds.ComponentLike{}).Where("component_id = ?", componentID).Count(&likesCount).Error

	var isLikedNow int64
	_ = r.db.Model(&ds.ComponentLike{}).Where("component_id = ? AND user_id = ?", componentID, userID).Count(&isLikedNow).Error

	return isLikedNow > 0, likesCount, nil
}

func (r *Repository) GetComponentWithCreator(id uint) (*ds.FullComponentSerializer, error) {
	var comp ds.ServerComponent
	err := r.db.Preload("Creator").Where("id = ? AND status != ?", id, "удален").First(&comp).Error
	if err != nil {
		return nil, err
	}

	return &ds.FullComponentSerializer{
		ServerComponent: comp,
		Creator:         comp.Creator,
	}, nil
}

func (r *Repository) GetFeedComponent(id int, next bool) (*ds.ServerComponent, int64, error) {
	var comp ds.ServerComponent

	if id == 0 {
		err := r.db.Where("status = ?", "опубликован").Order("id asc").Limit(1).First(&comp).Error
		if err != nil {
			return nil, 0, err
		}
	} else if next {
		err := r.db.Where("id > ? AND status = ?", id, "опубликован").Order("id asc").Limit(1).First(&comp).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = r.db.Where("status = ?", "опубликован").Order("id asc").Limit(1).First(&comp).Error
				if err != nil {
					return nil, 0, err
				}
			} else {
				return nil, 0, err
			}
		}
	} else {
		err := r.db.Where("id = ?", id).Limit(1).First(&comp).Error
		if err != nil {
			return nil, 0, err
		}
		if comp.Status == "удален" {
			return nil, 0, fmt.Errorf("данная услуга удалена и недоступна для просмотра")
		}
		if comp.Status != "опубликован" {
			return nil, 0, fmt.Errorf("услуга находится в статусе '%s'", comp.Status)
		}
	}

	var likesCount int64
	_ = r.db.Model(&ds.ComponentLike{}).Where("component_id = ?", comp.ID).Count(&likesCount).Error

	return &comp, likesCount, nil
}

func (r *Repository) GetPublishedCards(search string, minTDP int) ([]CardItemWithLikes, error) {
	var components []ds.ServerComponent

	tx := r.db.Where("status = ?", "опубликован")

	if search != "" {
		tx = tx.Where("name ILIKE ?", "%"+search+"%")
	}
	if minTDP > 0 {
		tx = tx.Where("tdp >= ?", minTDP)
	}

	err := tx.Order("id asc").Find(&components).Error
	if err != nil {
		return nil, err
	}

	var result []CardItemWithLikes
	for _, c := range components {
		var count int64
		_ = r.db.Model(&ds.ComponentLike{}).Where("component_id = ?", c.ID).Count(&count).Error
		result = append(result, CardItemWithLikes{
			Component: c,
			Likes:     count,
		})
	}

	return result, nil
}

func (r *Repository) GetUserDraft(userID uint) (*ds.ServerComponent, error) {
	var comp ds.ServerComponent
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").Limit(1).First(&comp).Error
	if err != nil {
		return nil, err
	}
	return &comp, nil
}

func (r *Repository) CreateDraft(userID uint, name string) (*ds.ServerComponent, error) {
	var count int64
	r.db.Model(&ds.ServerComponent{}).Where("creator_id = ? AND status = ?", userID, "черновик").Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("у вас уже есть активный черновик. Завершите его публикацию перед созданием нового")
	}

	invalidImageURL := "http://" + r.minioEndpoint + "/" + r.minioImageBucket + "/invalid.jpg"
	invalidVideoURL := "http://" + r.minioEndpoint + "/" + r.minioVideoBucket + "/invalid.mp4"

	draft := ds.ServerComponent{
		Name:          name,
		Description:   "",
		Status:        "черновик",
		ImageURL:      invalidImageURL,
		VideoURL:      invalidVideoURL,
		ComponentType: "",
		TDP:           0,
		CreatedAt:     time.Now(),
		CreatorID:     userID,
		FormedAt:      nil,
	}

	err := r.db.Create(&draft).Error
	if err != nil {
		return nil, err
	}

	return &draft, nil
}

func (r *Repository) PublishComponent(id uint, userID uint, description, componentType string, tdp int) error {
	var comp ds.ServerComponent
	err := r.db.Where("id = ? AND creator_id = ? AND status = ?", id, userID, "черновик").First(&comp).Error
	if err != nil {
		return fmt.Errorf("черновик не найден или уже опубликован")
	}

	now := time.Now()
	comp.Description = description
	comp.ComponentType = componentType
	comp.TDP = tdp
	comp.Status = "опубликован"
	comp.FormedAt = &now

	return r.db.Save(&comp).Error
}

func (r *Repository) DeleteComponent(id uint) error {
	query := "UPDATE server_components SET status = $1 WHERE id = $2"
	res := r.db.Exec(query, "удален", id)
	if res.Error != nil {
		return fmt.Errorf("ошибка SQL при логическом удалении компонента %d: %w", id, res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("компонент с id %d не найден", id)
	}
	return nil
}
