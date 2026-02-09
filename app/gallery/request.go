package gallery

type GalleryRequest struct {
	Title       string   `json:"title" form:"title" binding:"required,min=3,max=200"`
	Category    Category `json:"category" form:"category" binding:"required,oneof=photography painting sculpture digital mixed"`
	Description string   `json:"description" form:"description" binding:"required,min=10,max=1000"`
	Image       string   `json:"image" form:"image_url"` // Optional if uploading file
	Status      Status   `json:"status" form:"status" binding:"omitempty,oneof=active inactive archived"`
}

type UpdateGalleryRequest struct {
	Title       string   `json:"title" form:"title" binding:"omitempty,min=3,max=200"`
	Category    Category `json:"category" form:"category" binding:"omitempty,oneof=photography painting sculpture digital mixed"`
	Description string   `json:"description" form:"description" binding:"omitempty,min=10,max=1000"`
	Image       string   `json:"image" form:"image_url"`
	Status      Status   `json:"status" form:"status" binding:"omitempty,oneof=active inactive archived"`
}

type GalleryQueryParams struct {
	Category Category `form:"category" binding:"omitempty,oneof=photography painting sculpture digital mixed"`
	Status   Status   `form:"status" binding:"omitempty,oneof=active inactive archived"`
	Cursor   string   `form:"cursor" binding:"omitempty"`
	Limit    int      `form:"limit" binding:"omitempty,min=1,max=100"`
}
