package ds

import (
	"time"
)

type ServerComponent struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string     `gorm:"type:varchar(100);not null" json:"name"`
	Description   string     `gorm:"type:text;not null;default:''" json:"description"`
	Status        string     `gorm:"type:varchar(20);not null;default:'черновик'" json:"status"`
	ImageURL      string     `gorm:"type:varchar(255);not null;default:''" json:"image_url"`
	VideoURL      string     `gorm:"type:varchar(255);not null;default:''" json:"video_url"`
	ComponentType string     `gorm:"type:varchar(50);not null;default:''" json:"component_type"`
	TDP           int        `gorm:"type:int;not null;default:0" json:"tdp"`
	CreatedAt     time.Time  `gorm:"not null" json:"created_at"`
	CreatorID     uint       `gorm:"not null" json:"creator_id"`
	FormedAt      *time.Time `gorm:"type:timestamp" json:"formed_at"`

	Creator User `gorm:"foreignKey:CreatorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
}

func (ServerComponent) TableName() string {
	return "server_components"
}

type FullComponentSerializer struct {
	ServerComponent
	Creator User `json:"creator"`
}

type ComponentListItem struct {
	ServerComponent
	IsCreator  int   `json:"is_creator"`
	LikesCount int64 `json:"likes_count"`
}

type ComponentFeedItem struct {
	ServerComponent
	IsLiked    int   `json:"is_liked"`
	LikesCount int64 `json:"likes_count"`
}

type PublishReq struct {
	Description   string `json:"description"`
	ComponentType string `json:"component_type"`
	TDP           int    `json:"tdp"`
}

type LikeReq struct {
	Like int `json:"like"`
}
