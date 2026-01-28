package postgre_pkg

import (
	image "github.com/Vilamuzz/yota-backend/app/Image"
	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/donation"
	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/Vilamuzz/yota-backend/app/user"
)

func GetAllModels() []interface{} {
	return []interface{}{
		&user.User{},
		&auth.PasswordResetToken{},
		&donation.Donation{},
		&news.News{},
		&social_program.SocialProgram{},
		&image.Image{},
	}
}
