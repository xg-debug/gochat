package handler

import (
	"net/http"

	"gochat/internal/service"

	"github.com/gin-gonic/gin"
)

func UploadAvatar(c *gin.Context) {
	url, err := service.UploadService.UploadAvatar(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func UploadChatImage(c *gin.Context) {
	url, err := service.UploadService.UploadChatImage(c, int64(c.GetUint64("user_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func UploadChatFile(c *gin.Context) {
	url, err := service.UploadService.UploadChatFile(c, int64(c.GetUint64("user_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func UploadChatAudio(c *gin.Context) {
	url, err := service.UploadService.UploadChatAudio(c, int64(c.GetUint64("user_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func UploadGroupAvatar(c *gin.Context) {
	url, err := service.UploadService.UploadGroupAvatar(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}
