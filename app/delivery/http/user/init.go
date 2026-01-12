package user_http

import (
	"github.com/Vilamuzz/yota-backend/app/delivery/http/middleware"
	"github.com/Vilamuzz/yota-backend/domain"
	"github.com/gin-gonic/gin"
)

type routeUser struct {
	Usecase    domain.UserAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.AppMiddleware
}

func NewRouteUser(usecase domain.UserAppUsecase, ginEngine *gin.Engine, middleware middleware.AppMiddleware) {
	handler := &routeUser{
		Usecase:    usecase,
		Route:      ginEngine.Group("/user"),
		Middleware: middleware,
	}

	handler.handleAuthRoute("/auth")
}
