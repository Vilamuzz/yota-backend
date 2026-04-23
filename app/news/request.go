package news

import (
	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type NewsRequest struct {
	Title    string               `json:"title" form:"title"`
	Category media.MediaCategory  `json:"category" form:"category"`
	Content  string               `json:"content" form:"content"`
	Image    string               `json:"image" form:"image"` // Optional if uploading file
	Status   media.MediaStatus    `json:"status" form:"status"`
	Media    []media.MediaRequest `json:"media"`
}

type UpdateNewsRequest struct {
	Title    string               `json:"title" form:"title"`
	Category media.MediaCategory  `json:"category" form:"category"`
	Content  string               `json:"content" form:"content"`
	Image    string               `json:"image" form:"image"`
	Status   media.MediaStatus    `json:"status" form:"status"`
	Media    []media.MediaRequest `json:"media"`
}

type NewsQueryParams struct {
	Category media.MediaCategory `form:"category"`
	Status   media.MediaStatus   `form:"status"`
	pkg.CursorPagination
}
