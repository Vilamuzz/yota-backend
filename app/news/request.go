package news

type NewsRequest struct {
	Title    string   `json:"title" form:"title" binding:"required,min=5,max=200"`
	Category Category `json:"category" form:"category" binding:"required,oneof=general event announcement donation social"`
	Content  string   `json:"content" form:"content" binding:"required,min=50"`
	Image    string   `json:"image" form:"image"` // Optional if uploading file
	Status   Status   `json:"status" form:"status" binding:"omitempty,oneof=draft published archived"`
}

type UpdateNewsRequest struct {
	Title    string   `json:"title" form:"title" binding:"omitempty,min=5,max=200"`
	Category Category `json:"category" form:"category" binding:"omitempty,oneof=general event announcement donation social"`
	Content  string   `json:"content" form:"content" binding:"omitempty,min=50"`
	Image    string   `json:"image" form:"image"`
	Status   Status   `json:"status" form:"status" binding:"omitempty,oneof=draft published archived"`
}

type NewsQueryParams struct {
	Category Category `form:"category" binding:"omitempty,oneof=general event announcement donation social"`
	Status   Status   `form:"status" binding:"omitempty,oneof=draft published archived"`
	Cursor   string   `form:"cursor" binding:"omitempty"`
	Limit    int      `form:"limit" binding:"omitempty,min=1,max=100"`
}
