package auth

import "sync"

const CurrentUserID uint = 1

type SessionManager struct {
	userID uint
}

var (
	instance *SessionManager
	once     sync.Once
)

func GetCurrentUserID() uint {
	once.Do(func() {
		instance = &SessionManager{
			userID: CurrentUserID,
		}
	})
	return instance.userID
}
