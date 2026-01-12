package superadmin_http

import (
	"github.com/Vilamuzz/yota-backend/app/delivery/http/middleware"
	"github.com/Vilamuzz/yota-backend/domain"
	"github.com/gin-gonic/gin"
)

type routeSuperadmin struct {
	usecase    domain.SuperadminAppUsecase
	route      *gin.RouterGroup
	middleware middleware.AppMiddleware
}

func NewRouteSuperadmin(usecase domain.SuperadminAppUsecase, ginEngine *gin.Engine, middleware middleware.AppMiddleware) {
	handler := &routeSuperadmin{
		usecase:    usecase,
		route:      ginEngine.Group("/superadmin"),
		middleware: middleware,
	}
	handler.handleAuthRoute("/auth")
}
