package news

type NewsRequest struct {
	Title    string   `json:"title" binding:"required,min=5,max=200"`
	Category Category `json:"category" binding:"required,oneof=general event announcement donation social"`
	Content  string   `json:"content" binding:"required,min=50"`
	Image    string   `json:"image" binding:"omitempty,url"`
	Status   Status   `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type UpdateNewsRequest struct {
	Title    string   `json:"title" binding:"omitempty,min=5,max=200"`
	Category Category `json:"category" binding:"omitempty,oneof=general event announcement donation social"`
	Content  string   `json:"content" binding:"omitempty,min=50"`
	Image    string   `json:"image" binding:"omitempty,url"`
	Status   Status   `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type NewsQueryParams struct {
	Category Category `form:"category" binding:"omitempty,oneof=general event announcement donation social"`
	Status   Status   `form:"status" binding:"omitempty,oneof=draft published archived"`
	Cursor   string   `form:"cursor" binding:"omitempty"`
	Limit    int      `form:"limit" binding:"omitempty,min=1,max=100"`
}
