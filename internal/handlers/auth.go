package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
	"url_shortenn/internal/config"
	"url_shortenn/internal/models"
)

type AuthHandler struct {
	db     *sqlx.DB
	config *config.JWTConfig
}

func NewAuthHandler(db *sqlx.DB, jwtConfig *config.JWTConfig) *AuthHandler {
	return &AuthHandler{db: db, config: jwtConfig}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Register validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log input password
	log.Printf("Register - Email: %s, Input password: %s", req.Email, req.Password)

	// Kiểm tra email tồn tại
	var exists bool
	err := h.db.QueryRowContext(c, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check email"})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Log hashed password
	log.Printf("Register - Hashed password: %s", string(hashedPassword))

	query := "INSERT INTO users (email, password) VALUES ($1, $2) returning id"

	var userId int

	err = h.db.QueryRowContext(c, query, req.Email, string(hashedPassword)).Scan(&userId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert user"})
		return
	}

	token, err := h.generateToken(userId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})

}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Login binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.db.GetContext(c, &user, "SELECT * FROM users WHERE email = $1", req.Email)
	
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Check if user exists and has a password (not Google login)
	if user.GoogleID.Valid && user.GoogleID.String != "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "please login with Google"})
		return
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("Password comparison failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) generateToken(userId int) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userId),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // Explicitly set to 24 hours
	}
	
	// Add debug logging
	log.Printf("Generating token with claims: %+v", claims)
	log.Printf("Using secret key: %s", h.config.SecretKey)
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}
	return tokenString, nil
}




