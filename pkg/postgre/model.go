package postgre_pkg

import (
	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/user"
)

func GetAllModels() []interface{} {
	return []interface{}{
		&user.User{},
		&auth.PasswordResetToken{},
	}
}
