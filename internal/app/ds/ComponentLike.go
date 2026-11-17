package ds

type ComponentLike struct {
	ID          uint `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint `gorm:"not null;uniqueIndex:idx_user_component_like" json:"user_id"`
	ComponentID uint `gorm:"not null;uniqueIndex:idx_user_component_like" json:"component_id"`

	User            User            `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	ServerComponent ServerComponent `gorm:"foreignKey:ComponentID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
}

func (ComponentLike) TableName() string {
	return "component_likes"
}
