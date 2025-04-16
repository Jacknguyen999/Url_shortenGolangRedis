package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
	"url_shortenn/internal/config"
)

type OAuthHandler struct {
	db          *sqlx.DB
	config      *config.OAuthConfig
	oauthConfig *oauth2.Config
	authHandler *AuthHandler
}
type GoogleUserInfor struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	GoogleID string `json:"sub"`
}

func NewOAuthHandler(db *sqlx.DB, cfg *config.OAuthConfig, authHandler *AuthHandler) *OAuthHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &OAuthHandler{
		db:          db,
		config:      cfg,
		oauthConfig: oauthConfig,
		authHandler: authHandler,
	}
}

func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	log.Printf("OAuth Config: %+v", h.oauthConfig)
	log.Printf("Redirect URL: %s", h.config.RedirectURL)
	
	url := h.oauthConfig.AuthCodeURL("state-token")
	log.Printf("Auth URL: %s", url)
	
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	log.Printf("Callback received with code: %s", code)
	
	token, err := h.oauthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
		return
	}
	client := h.oauthConfig.Client(c, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfor GoogleUserInfor

	if err := json.NewDecoder(resp.Body).Decode(&userInfor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode user info"})
		return
	}

	var userId int
	err = h.db.QueryRowContext(c, "SELECT id FROM users WHERE google_id = $1", userInfor.GoogleID).Scan(&userId)

	if err != nil {
		query := "INSERT INTO users (email, google_id, name,avatar_url,password) VALUES ($1, $2, $3, $4, '') returning id"
		err = h.db.QueryRowContext(
			c, query, userInfor.Email,
			userInfor.GoogleID,
			userInfor.Name,
			userInfor.Picture,
		).Scan(&userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert user"})
			return
		}
	}
	jwtToken, err := h.authHandler.generateToken(userId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": jwtToken,
		"user": gin.H{
			"id":         userId,
			"email":      userInfor.Email,
			"name":       userInfor.Name,
			"avatar_url": userInfor.Picture,
		},
	})

}
