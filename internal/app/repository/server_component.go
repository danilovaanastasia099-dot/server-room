package repository

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"server_components/internal/app/ds"
)

type CardItemWithLikes struct {
	Component ds.ServerComponent
	Likes     int64
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

	invalidImageURL := "http://localhost:9000/server-images/invalid.jpg"
	invalidVideoURL := "http://localhost:9000/server-videos/invalid.mp4"

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
