package http

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/vele/temp_test_repo/internal/transport/http/handler"
	"github.com/vele/temp_test_repo/internal/transport/http/middleware"
)

type RouterDeps struct {
	UserHandler *handler.UserHandler
	AuthHandler *handler.AuthHandler
	Auth        *middleware.Auth
	Logger      *logrus.Logger
}

func NewRouter(deps RouterDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	if deps.Logger != nil {
		router.Use(middleware.RequestLogger(deps.Logger))
	}
	router.POST("/auth/login", deps.AuthHandler.Login)

	api := router.Group("/api/v1")
	api.Use(deps.Auth.Handler())
	deps.UserHandler.RegisterRoutes(api)

	return router
}
