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

func NewRouteSuperadmin(usecase domain.SuperadminAppUsecase, router *gin.RouterGroup, middleware middleware.AppMiddleware) {
	handler := &routeSuperadmin{
		usecase:    usecase,
		route:      router.Group("/superadmin"),
		middleware: middleware,
	}
	handler.handleAuthRoute("/auth")
}
