package fileserver

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/gin-gonic/gin"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	cachePath   string = "data/cache"
	imageDBPath string = "data/image"
	audioDBPath string = "data/audio"
	videoDBPath string = "data/video"
)

// FileServer 文件服务器
type FileServer struct {
	// URL 文件服务器 URL
	URL string
	// UseLocalFileServer 是否使用本地文件服务器
	UseLocalFileServer bool
}

// CacheData 缓存数据
type CacheData struct {
	FileInfo string // 文件信息
	TTL      int64  // 过期时间
	Time     int64  // 保存时间
}

// FileCacheDB 文件缓存数据库
type FileCacheDB struct {
	DB *leveldb.DB
	mu sync.Mutex
}

var (
	instance        *FileServer
	imageDBInstance *FileCacheDB
	audioDBInstance *FileCacheDB
	videoDBInstance *FileCacheDB
)

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

// startImageDB 启动图片数据库
func startImageDB() error {
	// 创建或打开图片缓存数据库
	db, err := leveldb.OpenFile(imageDBPath, nil)
	if err != nil {
		return err
	}

	imageDBInstance = &FileCacheDB{
		DB: db,
	}

	return nil
}

// startAudioDB 启动音频数据库
func startAudioDB() error {
	// 创建或打开音频缓存数据库
	db, err := leveldb.OpenFile(audioDBPath, nil)
	if err != nil {
		return err
	}

	audioDBInstance = &FileCacheDB{
		DB: db,
	}

	return nil
}

// startVideoDB 启动视频数据库
func startVideoDB() error {
	// 创建或打开视频缓存数据库
	db, err := leveldb.OpenFile(videoDBPath, nil)
	if err != nil {
		return err
	}

	videoDBInstance = &FileCacheDB{
		DB: db,
	}

	return nil
}

// getCache 获取缓存数据
func (cacheDB *FileCacheDB) getCache(key string) (string, bool) {
	cacheDB.mu.Lock()
	defer cacheDB.mu.Unlock()

	// 从数据库中获取缓存数据
	data, err := cacheDB.DB.Get([]byte(key), nil)
	if err != nil {
		return "", false
	}

	// 解码缓存数据
	buffer := *bytes.NewBuffer(data)
	decoder := gob.NewDecoder(&buffer)

	var cache CacheData
	err = decoder.Decode(&cache)
	if err != nil {
		return "", false
	}

	// 如果是可以长期使用的则直接返回
	if cache.TTL == 0 {
		return cache.FileInfo, true
	}

	// 获取当前时间，检查是否过期
	now := time.Now().Unix()
	if now > cache.Time+cache.TTL {
		// 过期则删除缓存数据
		_ = cacheDB.DB.Delete([]byte(key), nil)
		return "", false
	}

	return cache.FileInfo, true
}

// saveCache 保存缓存数据
func (cacheDB *FileCacheDB) saveCache(key string, fileInfo string, ttl int64) error {
	cacheDB.mu.Lock()
	defer cacheDB.mu.Unlock()

	// 获取当前时间
	now := time.Now().Unix()

	// 编码缓存数据
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(CacheData{
		FileInfo: fileInfo,
		TTL:      ttl,
		Time:     now,
	})
	if err != nil {
		return err
	}

	// 保存缓存数据
	err = cacheDB.DB.Put([]byte(key), buffer.Bytes(), nil)
	if err != nil {
		return err
	}

	return nil
}

// GetImageCache 获取图片缓存
func GetImageCache(key string) (string, bool) {
	if imageDBInstance == nil {
		if err := startImageDB(); err != nil {
			return "", false
		}
	}

	return imageDBInstance.getCache(key)
}

// SaveImageCache 保存图片缓存
func SaveImageCache(key string, fileInfo string, ttl int64) error {
	if imageDBInstance == nil {
		if err := startImageDB(); err != nil {
			return err
		}
	}

	return imageDBInstance.saveCache(key, fileInfo, ttl)
}

// GetAudioCache 获取音频缓存
func GetAudioCache(key string) (string, bool) {
	if audioDBInstance == nil {
		if err := startAudioDB(); err != nil {
			return "", false
		}
	}

	return audioDBInstance.getCache(key)
}

// SaveAudioCache 保存音频缓存
func SaveAudioCache(key string, fileInfo string, ttl int64) error {
	if audioDBInstance == nil {
		if err := startAudioDB(); err != nil {
			return err
		}
	}

	return audioDBInstance.saveCache(key, fileInfo, ttl)
}

// GetVideoCache 获取视频缓存
func GetVideoCache(key string) (string, bool) {
	if videoDBInstance == nil {
		if err := startVideoDB(); err != nil {
			return "", false
		}
	}

	return videoDBInstance.getCache(key)
}

// SaveVideoCache 保存视频缓存
func SaveVideoCache(key string, fileInfo string, ttl int64) error {
	if videoDBInstance == nil {
		if err := startVideoDB(); err != nil {
			return err
		}
	}

	return videoDBInstance.saveCache(key, fileInfo, ttl)
}

// GetHash 获取文件哈希值
func GetHash(data []byte) string {
	return hash(data)
}
