package postgre_pkg

import (
	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/donation"
	"github.com/Vilamuzz/yota-backend/app/gallery"
	"github.com/Vilamuzz/yota-backend/app/image"
	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/Vilamuzz/yota-backend/app/user"
)

func GetAllModels() []interface{} {
	return []interface{}{
		&user.User{},
		&auth.PasswordResetToken{},
		&auth.EmailVerificationToken{},
		&donation.Donation{},
		&news.News{},
		&gallery.Gallery{},
		&image.Image{},
		&social_program.SocialProgram{},
	}
}
