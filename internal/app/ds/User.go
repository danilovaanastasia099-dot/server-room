package ds

type User struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Login       string `gorm:"type:varchar(50);not null;unique" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}

func (User) TableName() string {
	return "users"
}

type UserComponents struct {
	User       User              `json:"user"`
	Components []ServerComponent `json:"components"`
}

type RegisterReq struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginReq struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}
