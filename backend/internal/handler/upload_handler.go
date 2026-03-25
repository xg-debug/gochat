package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *App) UploadAvatar(c *gin.Context) {
	url, err := h.Upload.UploadAvatar(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *App) UploadChatImage(c *gin.Context) {
	url, err := h.Upload.UploadChatImage(c, currentUserID(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *App) UploadChatFile(c *gin.Context) {
	url, err := h.Upload.UploadChatFile(c, currentUserID(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *App) UploadChatAudio(c *gin.Context) {
	url, err := h.Upload.UploadChatAudio(c, currentUserID(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *App) UploadGroupAvatar(c *gin.Context) {
	url, err := h.Upload.UploadGroupAvatar(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}
