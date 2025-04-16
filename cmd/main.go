package main

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"url_shortenn/internal/config"
	"url_shortenn/internal/database"
	"url_shortenn/internal/handlers"
	"url_shortenn/internal/routes"
	"url_shortenn/internal/service"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Add debug logging for JWT configuration
	log.Printf("JWT Configuration loaded: SecretKey length: %d, ExpireTime: %v", 
		len(cfg.JWT.SecretKey), cfg.JWT.TokenExpireTime)

	db, err := database.DB(&cfg.Database)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.InitSchema(db); err != nil {
		log.Fatal("Failed to initialize database schema:", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	defer rdb.Close()

	urlService := service.NewURLService(db, rdb)

	urlHandler := handlers.NewURLHandler(urlService)

	authhandler := handlers.NewAuthHandler(db, &cfg.JWT)

	log.Printf("OAuth Config loaded: %+v", cfg.OAuth)
	oauthHandler := handlers.NewOAuthHandler(db, &cfg.OAuth, authhandler)

	r := gin.Default()

	routes.Routes(r, urlHandler, authhandler, oauthHandler, &cfg.JWT)

	log.Printf("Listening on %s", cfg.Server.Port)

	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatal(err)
	}

}

