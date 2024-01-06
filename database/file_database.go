package database

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	imageDBPath string = "data/image"
	audioDBPath string = "data/audio"
	videoDBPath string = "data/video"
)

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
	imageDBInstance *FileCacheDB
	audioDBInstance *FileCacheDB
	videoDBInstance *FileCacheDB
)

// StartFileDB 启动文件数据库
func StartFileDB() {
	log.Info("正在启动数据库...")

	// 启动图片数据库
	if err := startImageDB(); err != nil {
		log.Error("启动图片数据库失败: ", err)
	} else {
		log.Info("图片数据库已启动")
	}

	// 启动音频数据库
	if err := startAudioDB(); err != nil {
		log.Error("启动音频数据库失败: ", err)
	} else {
		log.Info("音频数据库已启动")
	}

	// 启动视频数据库
	if err := startVideoDB(); err != nil {
		log.Error("启动视频数据库失败: ", err)
	} else {
		log.Info("视频数据库已启动")
	}
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
		return "", false
	}

	return imageDBInstance.getCache(key)
}

// SaveImageCache 保存图片缓存
func SaveImageCache(key string, fileInfo string, ttl int64) error {
	if imageDBInstance == nil {
		return fmt.Errorf("图片数据库未启动")
	}

	return imageDBInstance.saveCache(key, fileInfo, ttl)
}

// GetAudioCache 获取音频缓存
func GetAudioCache(key string) (string, bool) {
	if audioDBInstance == nil {
		return "", false
	}

	return audioDBInstance.getCache(key)
}

// SaveAudioCache 保存音频缓存
func SaveAudioCache(key string, fileInfo string, ttl int64) error {
	if audioDBInstance == nil {
		return fmt.Errorf("音频数据库未启动")
	}

	return audioDBInstance.saveCache(key, fileInfo, ttl)
}

// GetVideoCache 获取视频缓存
func GetVideoCache(key string) (string, bool) {
	if videoDBInstance == nil {
		return "", false
	}

	return videoDBInstance.getCache(key)
}

// SaveVideoCache 保存视频缓存
func SaveVideoCache(key string, fileInfo string, ttl int64) error {
	if videoDBInstance == nil {
		return fmt.Errorf("视频数据库未启动")
	}

	return videoDBInstance.saveCache(key, fileInfo, ttl)
}
