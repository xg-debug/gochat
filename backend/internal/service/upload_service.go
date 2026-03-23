package service

import (
	"errors"
	"fmt"
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
	MaxUploadSize      = 5 * 1024 * 1024
	AvatarUploadPath   = "uploads/avatars"
	ChatImagePath      = "uploads/chat"
	ChatImageMaxSize   = 5 * 1024 * 1024
	ChatFilePath       = "uploads/files"
	ChatFileMaxSize    = 20 * 1024 * 1024
	ChatAudioPath      = "uploads/audio"
	ChatAudioMaxSize   = 10 * 1024 * 1024
	GroupAvatarPath    = "uploads/groups"
	GroupAvatarMaxSize = 5 * 1024 * 1024
)

type uploadService struct{}

var UploadService = &uploadService{}

func (s *uploadService) UploadAvatar(c *gin.Context) (string, error) {
	return saveUploaded(c, "file", AvatarUploadPath, MaxUploadSize, []string{".jpg", ".jpeg", ".png", ".gif"}, "/uploads/avatars", false)
}

func (s *uploadService) UploadChatImage(c *gin.Context, userID int64) (string, error) {
	url, err := saveUploaded(c, "file", ChatImagePath, ChatImageMaxSize, []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}, "/uploads/chat", false)
	if err != nil {
		return "", err
	}
	s.saveFileRecord(userID, c, url, "image")
	return url, nil
}

func (s *uploadService) UploadChatFile(c *gin.Context, userID int64) (string, error) {
	url, err := saveUploaded(c, "file", ChatFilePath, ChatFileMaxSize, nil, "/uploads/files", true)
	if err != nil {
		return "", err
	}
	s.saveFileRecord(userID, c, url, "file")
	return url, nil
}

func (s *uploadService) UploadChatAudio(c *gin.Context, userID int64) (string, error) {
	url, err := saveUploaded(c, "file", ChatAudioPath, ChatAudioMaxSize, []string{".mp3", ".wav", ".ogg", ".m4a", ".webm"}, "/uploads/audio", false)
	if err != nil {
		return "", err
	}
	s.saveFileRecord(userID, c, url, "audio")
	return url, nil
}

func (s *uploadService) UploadGroupAvatar(c *gin.Context) (string, error) {
	return saveUploaded(c, "file", GroupAvatarPath, GroupAvatarMaxSize, []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}, "/uploads/groups", false)
}

func (s *uploadService) saveFileRecord(userID int64, c *gin.Context, url, fileType string) {
	if userID <= 0 {
		return
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		return
	}
	_ = dbConn.Create(&model.File{UserID: userID, FileName: file.Filename, FileURL: url, FileSize: file.Size, FileType: fileType, CreatedAt: time.Now()}).Error
}

func saveUploaded(c *gin.Context, field, dstDir string, maxSize int64, exts []string, urlPrefix string, allowAnyExt bool) (string, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return "", errors.New("获取文件失败")
	}
	if file.Size > maxSize {
		return "", fmt.Errorf("文件大小不能超过%dMB", maxSize/1024/1024)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowAnyExt {
		ok := false
		for _, allowed := range exts {
			if ext == allowed {
				ok = true
				break
			}
		}
		if !ok {
			return "", errors.New("文件格式不支持")
		}
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return "", errors.New("创建上传目录失败")
	}
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(dstDir, newFileName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", errors.New("保存文件失败")
	}
	return fmt.Sprintf("%s/%s", urlPrefix, newFileName), nil
}
