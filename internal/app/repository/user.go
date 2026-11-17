package repository

import (
	"fmt"

	"server_components/internal/app/ds"
)

func (r *Repository) RegisterUser(login, password string) (*ds.User, error) {
	if login == "" || password == "" {
		return nil, fmt.Errorf("логин и пароль обязательны для заполнения")
	}

	var count int64
	r.db.Model(&ds.User{}).Where("login = ?", login).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("пользователь с логином '%s' уже существует", login)
	}

	user := ds.User{
		Login:       login,
		Password:    password,
		IsModerator: false,
	}

	err := r.db.Create(&user).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return &user, nil
}

func (r *Repository) GetUserByID(id uint) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserWithComponents(userID uint) (*ds.UserComponents, error) {
	var user ds.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("пользователь с id %d не найден: %w", userID, err)
	}

	var components []ds.ServerComponent
	err = r.db.Where("creator_id = ? AND status != ?", userID, "удален").Order("id asc").Find(&components).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка получения услуг пользователя: %w", err)
	}

	return &ds.UserComponents{
		User:       user,
		Components: components,
	}, nil
}
