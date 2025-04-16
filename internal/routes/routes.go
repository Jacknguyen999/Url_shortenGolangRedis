package routes

import (
	"time"

	"url_shortenn/internal/config"
	"url_shortenn/internal/handlers"
	"url_shortenn/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine, urlHandler *handlers.URLHandler, authHandler *handlers.AuthHandler, OAuthHandler *handlers.OAuthHandler, jwtConfig *config.JWTConfig) {

	// Rate limit riêng cho auth routes
	auth := r.Group("/auth")
	auth.Use(middleware.RateLimit(urlHandler.UrlService.Redis, 10, time.Minute))
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)

		auth.GET("/google/login", OAuthHandler.GoogleLogin)
		auth.GET("/google/callback", OAuthHandler.GoogleCallback)
	}

	// Rate limit riêng cho public routes
	r.GET("/:shortURL", middleware.RateLimit(urlHandler.UrlService.Redis, 1000, time.Hour), urlHandler.Redirect)

	// Rate limit riêng cho API routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(jwtConfig))
	api.Use(middleware.RateLimit(urlHandler.UrlService.Redis, 100, time.Hour))
	{
		api.POST("/shorten", urlHandler.Short)
		api.GET("/urls", urlHandler.GetUserURL)
		api.DELETE("/url/:id", urlHandler.DeleteURL)
		api.PUT("/url/:id", urlHandler.UpdateURL)
	}
}
