package admin_http

import (
	"github.com/Vilamuzz/yota-backend/app/delivery/http/middleware"
	"github.com/Vilamuzz/yota-backend/domain"
	"github.com/gin-gonic/gin"
)

type routeAdmin struct {
	usecase    domain.AdminAppUsecase
	route      *gin.RouterGroup
	middleware middleware.AppMiddleware
}

func NewRouteAdmin(usecase domain.AdminAppUsecase, ginEngine *gin.Engine, middleware middleware.AppMiddleware) {
	handler := &routeAdmin{
		usecase:    usecase,
		route:      ginEngine.Group("/admin"),
		middleware: middleware,
	}
	handler.handleAuthRoute("/auth")
}
