package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *URLHandler) Redirect(c *gin.Context) {
	shortURL := c.Param("shortURL")

	longURL, err := h.UrlService.GetLongURL(c.Request.Context(), shortURL)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusMovedPermanently, longURL)
}
