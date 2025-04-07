package routes

import (
	"github.com/gin-gonic/gin"
	"time"
	"url_shortenn/internal/handlers"
	"url_shortenn/internal/middleware"
)

func Routes(r *gin.Engine, urlHandler *handlers.URLHandler, authHandler *handlers.AuthHandler) {

	// public
	r.Use(middleware.RateLimit(urlHandler.UrlService.Redis, 100, time.Hour))
	r.GET("/:shortURL", urlHandler.Redirect)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/shorten", urlHandler.Short)
		api.GET("/urls", urlHandler.GetUserURL)
		api.DELETE("/url/:id", urlHandler.DeleteURL)
		api.PUT("/url/:id", urlHandler.UpdateURL)
	}

}
