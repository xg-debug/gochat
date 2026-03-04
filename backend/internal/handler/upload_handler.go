package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gochat/internal/model"
	"gochat/internal/pkg/db"
)

const (
	MaxUploadSize     = 5 * 1024 * 1024 // 5MB
	AvatarUploadPath  = "uploads/avatars"
	ChatImagePath     = "uploads/chat"
	ChatImageMaxSize  = 5 * 1024 * 1024
	ChatFilePath      = "uploads/files"
	ChatFileMaxSize   = 20 * 1024 * 1024
	ChatAudioPath     = "uploads/audio"
	ChatAudioMaxSize  = 10 * 1024 * 1024
	GroupAvatarPath   = "uploads/groups"
	GroupAvatarMaxSize = 5 * 1024 * 1024
)

// UploadAvatar 上传头像
func UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}

	if file.Size > MaxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过5MB"})
		return
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持jpg, png, gif格式的图片"})
		return
	}

	// 确保目录存在
	if err := os.MkdirAll(AvatarUploadPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	// 生成新文件名
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(AvatarUploadPath, newFileName)

	// 保存文件
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 返回文件访问URL
	// 注意：这里假设静态资源服务已配置为 /uploads/avatars/ -> uploads/avatars/
	url := fmt.Sprintf("/uploads/avatars/%s", newFileName)
	
	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

// UploadChatImage 上传聊天图片
func UploadChatImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}

	if file.Size > ChatImageMaxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过5MB"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持jpg, png, gif, webp格式的图片"})
		return
	}

	if err := os.MkdirAll(ChatImagePath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(ChatImagePath, newFileName)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	url := fmt.Sprintf("/uploads/chat/%s", newFileName)

	userID := int64(c.GetUint64("user_id"))
	if userID > 0 {
		if dbConn := db.GetDB(); dbConn != nil {
			dbConn.Create(&model.File{
				UserID:    userID,
				FileName:  file.Filename,
				FileURL:   url,
				FileSize:  file.Size,
				FileType:  "image",
				CreatedAt: time.Now(),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

// UploadChatFile 上传聊天文件
func UploadChatFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}
	if file.Size > ChatFileMaxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过20MB"})
		return
	}
	if err := os.MkdirAll(ChatFilePath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(ChatFilePath, newFileName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	url := fmt.Sprintf("/uploads/files/%s", newFileName)
	userID := int64(c.GetUint64("user_id"))
	if userID > 0 {
		if dbConn := db.GetDB(); dbConn != nil {
			dbConn.Create(&model.File{
				UserID:    userID,
				FileName:  file.Filename,
				FileURL:   url,
				FileSize:  file.Size,
				FileType:  "file",
				CreatedAt: time.Now(),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// UploadChatAudio 上传语音
func UploadChatAudio(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}
	if file.Size > ChatAudioMaxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过10MB"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".mp3" && ext != ".wav" && ext != ".ogg" && ext != ".m4a" && ext != ".webm" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持mp3, wav, ogg, m4a, webm"})
		return
	}
	if err := os.MkdirAll(ChatAudioPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(ChatAudioPath, newFileName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	url := fmt.Sprintf("/uploads/audio/%s", newFileName)
	userID := int64(c.GetUint64("user_id"))
	if userID > 0 {
		if dbConn := db.GetDB(); dbConn != nil {
			dbConn.Create(&model.File{
				UserID:    userID,
				FileName:  file.Filename,
				FileURL:   url,
				FileSize:  file.Size,
				FileType:  "audio",
				CreatedAt: time.Now(),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// UploadGroupAvatar 上传群头像
func UploadGroupAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}
	if file.Size > GroupAvatarMaxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过5MB"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持jpg, png, gif, webp格式的图片"})
		return
	}
	if err := os.MkdirAll(GroupAvatarPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(GroupAvatarPath, newFileName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	url := fmt.Sprintf("/uploads/groups/%s", newFileName)
	c.JSON(http.StatusOK, gin.H{"url": url})
}
