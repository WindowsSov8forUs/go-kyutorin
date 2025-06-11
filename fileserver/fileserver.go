package fileserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/WindowsSov8forUs/glyccat/config"
	"github.com/WindowsSov8forUs/glyccat/log"
)

const (
	filePath       = "data/files"
	internalFormat = "internal:%s/%s/_tmp/%s"
	defaultTTL     = 7 * 24 * time.Hour // 默认文件有效期为 7 天
)

// FileServer 文件服务器
type FileServer struct {
	// version 协议版本
	version string
	// path Satori 服务器路径
	path string
	// URL 文件服务器 URL
	URL string
	// EnableLocalFileServer 是否使用本地文件服务器
	Enable bool
	// TTL 默认文件有效期
	TTL time.Duration
	// MetaDB 文件元数据数据库
	MetaDB *MetaDatabase
	// FileInfoDB 文件信息数据库
	FileInfoDB *FileInfoDatabase
}

var instance *FileServer

// StartFileServer 启动文件服务器
func StartFileServer(conf *config.Config) {
	log.Info("正在启动文件服务器...")

	// 去除可能的 / 结尾
	publicURL := conf.FileServer.ExternalURL
	publicURL = strings.TrimSuffix(publicURL, "/")

	// 确保文件服务器路径存在
	if err := os.MkdirAll(filePath, 0755); err != nil {
		log.Errorf("创建文件服务器目录失败: %s", err)
		instance = nil
		return
	}

	// 启动文件元数据数据库
	metaDB, err := StartMetaDB()
	if err != nil {
		log.Errorf("启动文件数据库失败: %s", err)
		instance = nil
		return
	}

	// 启动文件信息数据库
	fileInfoDB, err := StartFileInfoDB()
	if err != nil {
		log.Errorf("启动文件信息数据库失败: %s", err)
		instance = nil
		return
	}

	instance = &FileServer{
		version:    fmt.Sprintf("v%d", conf.Satori.Version),
		path:       conf.Satori.Path,
		URL:        publicURL,
		TTL:        time.Duration(conf.FileServer.TTL) * time.Second,
		Enable:     conf.FileServer.Enable,
		MetaDB:     metaDB,
		FileInfoDB: fileInfoDB,
	}

	// 清理过期文件
	metaDBCleanup()
	fileInfoDBCleanup()

	// 确保文件服务器对公网开放
	if instance.URL == "" {
		log.Warn("文件服务器未配置公网地址，可能导致无法向 QQ 开放平台上传文件。")
		instance.Enable = false
		instance.URL = fmt.Sprintf("127.0.0.1:%d", conf.Satori.Server.Port)
	}

	if instance.Enable {
		log.Infof("文件服务器已启动，公网 IP : %s", instance.URL)
	}
}

// metaDBCleanup 文件元数据数据库清理
func metaDBCleanup() {
	if instance == nil || !instance.Enable {
		return
	}

	log.Trace("正在清理文件数据库...")

	// 获取所有文件元数据
	files, err := instance.MetaDB.GetFileMetas()
	if err != nil {
		log.Tracef("获取文件元数据失败: %s", err)
		return
	}

	for _, meta := range files {
		// 启动文件清理定时器
		fileCleaner(meta)
	}

	log.Trace("文件数据库清理完成。")
}

// fileInfoDBCleanup 文件信息数据库清理
func fileInfoDBCleanup() {
	if instance == nil || !instance.Enable {
		return
	}

	log.Trace("正在清理文件信息数据库...")

	// 获取所有文件信息
	infos, err := instance.FileInfoDB.GetFileInfos()
	if err != nil {
		log.Tracef("获取文件信息失败: %s", err)
		return
	}

	for _, info := range infos {
		// 启动文件信息清理定时器
		fileInfoCleaner(info)
	}

	log.Trace("文件信息数据库清理完成。")
}

// CalculateFileIdent 计算文件标识符
func CalculateFileIdent(platform, userId string, file io.Reader) (string, error) {
	// 创建来源哈希
	fromIdent := platform + ":" + userId
	fromHash := sha256.Sum256([]byte(fromIdent))

	// 计算文件内容哈希，使用来源哈希作为 HMAC 密钥
	h := hmac.New(sha256.New, fromHash[:])
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	fileHash := hex.EncodeToString(h.Sum(nil))
	return fileHash, nil
}

// CalculateFileInfoIdent 计算文件信息标识符
func CalculateFileInfoIdent(targetId, src string) (string, error) {
	// 创建来源哈希
	fromHash := sha256.Sum256([]byte(targetId))

	// 计算 src 哈希，使用来源哈希作为 HMAC 密钥
	h := hmac.New(sha256.New, fromHash[:])
	if _, err := io.WriteString(h, src); err != nil {
		return "", err
	}
	fileInfoHash := hex.EncodeToString(h.Sum(nil))
	return fileInfoHash, nil
}

// SaveFile 保存文件并返回内部链接
func SaveFile(file io.Reader, platform, userId, name, fileType string) (*FileMetadata, error) {
	if instance == nil || !instance.Enable {
		return nil, fmt.Errorf("文件服务器未启用！")
	}

	// 确保文件目录存在
	if err := os.MkdirAll(filePath, 0755); err != nil {
		log.Errorf("创建文件目录失败: %s", err)
		return nil, err
	}

	// 读取文件内容到内存
	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Errorf("读取文件内容失败: %s", err)
		return nil, err
	}

	// 生成文件名
	fileName, err := CalculateFileIdent(platform, userId, strings.NewReader(string(fileContent)))
	if err != nil {
		log.Errorf("计算文件标识符失败: %s", err)
		return nil, err
	}

	// 保存数据
	path := filepath.Join(filePath, fileName)
	if err := os.WriteFile(path, fileContent, 0644); err != nil {
		log.Errorf("保存文件失败: %s", err)
		return nil, err
	}

	// 存储文件元数据
	meta := &FileMetadata{
		ID:          fileName,
		Name:        name,
		URL:         fmt.Sprintf(internalFormat, platform, userId, fileName),
		Path:        path,
		ContentType: fileType,
		CreateAt:    uint64(time.Now().Unix()),
		TTL:         uint64(instance.TTL.Seconds()),
	}
	if err := instance.MetaDB.SaveFileMeta(fileName, meta); err != nil {
		log.Errorf("保存文件元数据失败: %s", err)
	}

	// 启动文件清理定时器
	meta.cleanerTimer = fileCleaner(meta)

	return meta, nil
}

// SaveFileInfo 保存文件信息
func SaveFileInfo(targetId, src, fileInfo string, ttl uint64) (*FileInfo, error) {
	if instance == nil {
		return nil, fmt.Errorf("文件服务器未启用！")
	}

	// 生成文件信息标识符
	ident, err := CalculateFileInfoIdent(targetId, src)
	if err != nil {
		log.Errorf("计算文件信息标识符失败: %s", err)
		return nil, err
	}

	// 存储文件信息
	info := &FileInfo{
		ID:       ident,
		FileInfo: fileInfo,
		CreateAt: uint64(time.Now().Unix()),
		TTL:      ttl,
	}
	if err := instance.FileInfoDB.SaveFileInfo(ident, info); err != nil {
		log.Errorf("保存文件信息失败: %s", err)
		return nil, err
	}

	// 启动文件信息清理定时器
	info.cleanerTimer = fileInfoCleaner(info)

	return info, nil
}

// GetFile 获取文件内容
func GetFile(ident string) (*FileMetadata, error) {
	if instance == nil || !instance.Enable {
		return nil, fmt.Errorf("文件服务器未启用！")
	}

	// 获取文件元数据
	meta, err := instance.MetaDB.GetFileMeta(ident)
	if err != nil {
		log.Errorf("获取文件元数据失败: %s", err)
		return nil, err
	}

	// 检查文件是否存在
	filePath := filepath.Join(filePath, ident)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Errorf("文件不存在: %s", filePath)
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	return meta, nil
}

// GetFileInfo 获取文件信息
func GetFileInfo(ident string) (*FileInfo, error) {
	if instance == nil {
		return nil, fmt.Errorf("文件服务器未启用！")
	}

	// 获取文件信息
	info, err := instance.FileInfoDB.GetFileInfo(ident)
	if err != nil {
		log.Errorf("获取文件信息失败: %s", err)
		return nil, err
	}

	return info, nil
}

// DeleteFile 删除文件
func DeleteFile(ident string) error {
	if instance == nil || !instance.Enable {
		return nil
	}

	// 删除文件元数据
	if err := instance.MetaDB.DeleteFileMeta(ident); err != nil {
		log.Errorf("删除文件元数据失败: %s", err)
	}

	// 删除文件
	return os.Remove(filepath.Join(filePath, ident))
}

// DeleteFileInfo 删除文件信息
func DeleteFileInfo(ident string) error {
	if instance == nil {
		return nil
	}

	// 删除文件信息
	if err := instance.FileInfoDB.DeleteFileInfo(ident); err != nil {
		log.Errorf("删除文件信息失败: %s", err)
		return err
	}

	return nil
}

// fileCleaner 定期清理过期文件
func fileCleaner(meta *FileMetadata) *time.Timer {
	if meta.TTL == 0 {
		return nil // 没有设置过期时间
	}

	// 计算过期时间戳
	expireAt := meta.CreateAt + meta.TTL
	now := uint64(time.Now().Unix())
	if expireAt <= now {
		// 文件已过期，删除文件和元数据
		fileCleanerFunc(meta)
	}

	// 计算过期时长
	duration := time.Duration(expireAt-now) * time.Second
	// 启动定时器
	return time.AfterFunc(duration, func() {
		fileCleanerFunc(meta)
	})
}

// fileCleanerFunc 文件清理函数
func fileCleanerFunc(meta *FileMetadata) {
	if err := DeleteFile(meta.ID); err != nil {
		log.Errorf("清理文件失败: %s", err)
	}
}

// fileInfoCleaner 定期清理过期文件信息
func fileInfoCleaner(meta *FileInfo) *time.Timer {
	if meta.TTL == 0 {
		return nil // 没有设置过期时间
	}

	// 计算过期时间戳
	expireAt := meta.CreateAt + meta.TTL
	now := uint64(time.Now().Unix())
	if expireAt <= now {
		// 文件信息已过期，删除文件信息
		fileInfoCleanerFunc(meta)
	}

	// 计算过期时长
	duration := time.Duration(expireAt-now) * time.Second
	// 启动定时器
	return time.AfterFunc(duration, func() {
		fileInfoCleanerFunc(meta)
	})
}

// fileInfoCleanerFunc 文件信息清理函数
func fileInfoCleanerFunc(info *FileInfo) {
	if err := DeleteFileInfo(info.ID); err != nil {
		log.Errorf("清理文件信息失败: %s", err)
	}
}

// InternalURLPrefix 获取内部链接前缀
func InternalURLPrefix() string {
	if instance == nil || !instance.Enable {
		return ""
	}
	return fmt.Sprintf("http://%s%s/%s/proxy/", instance.URL, instance.path, instance.version)
}

// InternalURL 获取内部链接格式
func InternalURL(meta *FileMetadata) string {
	if instance == nil || !instance.Enable {
		return ""
	}
	return InternalURLPrefix() + meta.URL
}

// ParseInternalURL 解析内部链接
func ParseInternalURL(url string) (string, string, string, bool) {
	internalPattern := `^internal:([^/]+)/([^/]+)/(.+)$`
	re := regexp.MustCompile(internalPattern)
	matches := re.FindStringSubmatch(url)

	if len(matches) != 4 {
		return "", "", "", false
	}

	return matches[1], matches[2], matches[3], true
}

// GetPath 获取文件本地路径
func GetPath(path string) (string, error) {
	if instance == nil || !instance.Enable {
		return "", fmt.Errorf("文件服务器未启用！")
	}

	// 对于 _tmp 文件路径
	if strings.HasPrefix(path, "_tmp/") {
		// 提取文件 ident
		ident := strings.TrimPrefix(path, "_tmp/")
		if meta, err := GetFile(ident); err == nil {
			return meta.Path, nil
		} else {
			return "", err
		}
	}

	return "", fmt.Errorf("无效的文件路径: %s", path)
}
