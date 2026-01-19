package user

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     Role   `json:"role" binding:"omitempty,oneof=admin user"`
	Status   Status `json:"status" binding:"omitempty,oneof=active inactive banned"`
}
