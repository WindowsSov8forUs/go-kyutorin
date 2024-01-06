package fileserver

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/gin-gonic/gin"
)

const cachePath string = "data/cache"

// FileServer 文件服务器
type FileServer struct {
	// URL 文件服务器 URL
	URL string
	// UseLocalFileServer 是否使用本地文件服务器
	UseLocalFileServer bool
}

var instance *FileServer

// StartFileServer 启动文件服务器
func StartFileServer(conf *config.Config) {
	log.Info("正在启动文件服务器...")

	// 获取文件服务器路径
	fileServerPath := cachePath

	// 确保文件服务器路径存在
	if err := os.MkdirAll(fileServerPath, 0755); err != nil {
		log.Error("创建文件服务器目录失败: ", err)
		conf.FileServer.UseLocalFileServer = false
		return
	}

	// 创建文件服务器
	fileRouter := gin.Default()
	fileRouter.StaticFS("/files", gin.Dir(fileServerPath, true))

	instance = &FileServer{
		URL:                fmt.Sprintf("%s:%d/files", conf.FileServer.URL, conf.FileServer.Port),
		UseLocalFileServer: conf.FileServer.UseLocalFileServer,
	}

	log.Infof("文件服务器已在 %s 启动", instance.URL)

	// 启动文件服务器
	go func() {
		if err := fileRouter.Run(fmt.Sprintf("0.0.0.0:%d", conf.FileServer.Port)); err != nil {
			log.Error("文件服务器运行时出错: ", err)
			conf.FileServer.UseLocalFileServer = false
			instance = nil
			return
		}
	}()
}

// SaveFile 保存文件并返回 URL
func SaveFile(data []byte) string {
	if instance == nil || !instance.UseLocalFileServer {
		return ""
	}

	// 确保文件目录存在
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		log.Error("创建文件目录失败: ", err)
		return ""
	}

	// 保存数据
	filePath := filepath.Join(cachePath, hash(data))
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.Error("保存文件失败: ", err)
		return ""
	}

	// 返回 URL
	return "http://" + instance.URL + "/" + hash(data)
}

// hash 计算文件哈希值
func hash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// GetHash 获取文件哈希值
func GetHash(data []byte) string {
	return hash(data)
}
