package postgre_models

import (
	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/app/ambulance"
	"github.com/Vilamuzz/yota-backend/app/ambulance_history"
	"github.com/Vilamuzz/yota-backend/app/ambulance_request"
	"github.com/Vilamuzz/yota-backend/app/auth"
	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/Vilamuzz/yota-backend/app/donation_program_expense"
	"github.com/Vilamuzz/yota-backend/app/donation_program_transaction"
	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/app/foster_children"
	"github.com/Vilamuzz/yota-backend/app/foster_children_expense"
	"github.com/Vilamuzz/yota-backend/app/foster_children_transaction"
	"github.com/Vilamuzz/yota-backend/app/foundation_profile"
	"github.com/Vilamuzz/yota-backend/app/gallery"
	"github.com/Vilamuzz/yota-backend/app/log"
	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/Vilamuzz/yota-backend/app/news_comment"
	"github.com/Vilamuzz/yota-backend/app/prayer"
	"github.com/Vilamuzz/yota-backend/app/social_program"
	"github.com/Vilamuzz/yota-backend/app/social_program_expense"
	"github.com/Vilamuzz/yota-backend/app/social_program_invoice"
	"github.com/Vilamuzz/yota-backend/app/social_program_subscription"
	"github.com/Vilamuzz/yota-backend/app/social_program_transaction"
)

func GetAllModels() []interface{} {
	return []interface{}{
		&log.Log{},
		&account.Role{},
		&account.Account{},
		&account.UserProfile{},
		&account.AccountRole{},
		&auth.PasswordResetToken{},
		&auth.EmailVerificationToken{},
		&donation_program.DonationProgram{},
		&donation_program_transaction.DonationProgramTransaction{},
		&donation_program_expense.DonationProgramExpense{},
		&prayer.Prayer{},
		&prayer.PrayerAmen{},
		&prayer.PrayerReport{},
		&social_program.SocialProgram{},
		&social_program_subscription.SocialProgramSubscription{},
		&social_program_invoice.SocialProgramInvoice{},
		&social_program_transaction.SocialProgramTransaction{},
		&social_program_expense.SocialProgramExpense{},
		&foster_children.FosterChildren{},
		&foster_children.Achivement{},
		&foster_children.FosterChildrenCandidate{},
		&foster_children_expense.FosterChildrenExpense{},
		&foster_children_transaction.FosterChildrenTransaction{},
		&finance_record.FinanceRecord{},
		&foundation_profile.FoundationProfile{},
		&news.News{},
		&news_comment.NewsComment{},
		&news_comment.NewsCommentReport{},
		&gallery.Gallery{},
		&media.Media{},
		&ambulance.Ambulance{},
		&ambulance_request.AmbulanceRequest{},
		&ambulance_history.AmbulanceHistory{},
	}
}
