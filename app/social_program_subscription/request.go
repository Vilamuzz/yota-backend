package social_program_subscription

type CreateSocialProgramSubscriptionRequest struct {
	SocialProgramID string `json:"social_program_id"`
	UserID          string `json:"user_id"`
}

type UpdateSocialProgramSubscriptionRequest struct {
	Status Status `json:"status"`
}
