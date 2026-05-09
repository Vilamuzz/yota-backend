package news

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/app/media"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type NewsRequest struct {
	Title      string                  `json:"title" form:"title"`
	Category   media.MediaCategory     `json:"category" form:"category"`
	Content    string                  `json:"content" form:"content"`
	Status     media.MediaStatus       `json:"status" form:"status"`
	CoverImage *multipart.FileHeader   `form:"coverImage" swaggerignore:"true"`
	Metadata   []media.MediaMetadata   `form:"metadata"`
	Files      []*multipart.FileHeader `form:"files" swaggerignore:"true"`
}

type NewsQueryParams struct {
	Category media.MediaCategory `form:"category"`
	Status   media.MediaStatus   `form:"status"`
	pkg.CursorPagination
}
