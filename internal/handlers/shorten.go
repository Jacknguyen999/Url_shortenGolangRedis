package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"url_shortenn/internal/models"
	"url_shortenn/internal/service"
)

type URLHandler struct {
	UrlService *service.URLService
}

func NewURLHandler(urlService *service.URLService) *URLHandler {
	return &URLHandler{UrlService: urlService}
}

type ShortRequest struct {
	LongURl     string `json:"long_url" binding:"required,url"`
	CustomShort string `json:"custom_short,omitempty"`
}

func (h *URLHandler) Short(c *gin.Context) {
	var req ShortRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := c.GetString("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	url, err := h.UrlService.CreateURL(c.Request.Context(), req.LongURl, userID, req.CustomShort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *URLHandler) GetUserURL(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := strconv.Atoi(userIDStr)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}
	var urls []models.URL

	query := `SELECT * FROM urls WHERE user_id = $1 order by created_at desc`

	err = h.UrlService.DB.SelectContext(c.Request.Context(), &urls, query, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"urls": urls})
}
func (h *URLHandler) DeleteURL(c *gin.Context) {
	urlID := c.Param("id")

	userIDStr := c.GetString("user_id")

	userID, err := strconv.Atoi(userIDStr)

	result, err := h.UrlService.DB.ExecContext(c.Request.Context(),
		"DELETE FROM urls WHERE id = $1 AND user_id = $2", urlID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "url deleted"})
}

func (h *URLHandler) UpdateURL(c *gin.Context) {
	urlID := c.Param("id")
	userIDStr := c.GetString("user_id")
	userID, err := strconv.Atoi(userIDStr)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
	}

	var req struct {
		CustomShort string `json:"custom_short"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if custom url exist
	exist, _ := h.UrlService.ShortURLExists(c.Request.Context(), req.CustomShort)

	if exist {
		c.JSON(http.StatusBadRequest, gin.H{"message": "url already exists"})
		return
	}

	result, err := h.UrlService.DB.ExecContext(c.Request.Context(), "UPDATE urls set custom_short = $1 where id = $2 and user_id = $3", req.CustomShort, urlID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "url not found"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "url updated"})

}
