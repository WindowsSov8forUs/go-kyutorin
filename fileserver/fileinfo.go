package fileserver

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const fileInfoDatabasePath = "data/files/.fileinfo"

// FileInfo 文件信息
type FileInfo struct {
	ID           string      `json:"id"`        // 文件唯一标识 ID
	FileInfo     string      `json:"file_info"` // 开放平台返回的文件信息
	CreateAt     uint64      `json:"create_at"` // 文件创建时间戳
	TTL          uint64      `json:"ttl"`       // 文件有效时间
	cleanerTimer *time.Timer `json:"-"`         // 清理定时器，用于定时删除过期文件
}

// MarshalBinary 序列化文件信息为二进制
func (m *FileInfo) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary 反序列化二进制数据为文件信息
func (m *FileInfo) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	if err := json.NewDecoder(buf).Decode(m); err != nil {
		return err
	}
	return nil
}

// FileInfoDatabase 文件信息数据库
type FileInfoDatabase struct {
	DB *leveldb.DB
	mu sync.Mutex
}

// StartFileInfoDB 启动文件信息数据库
func StartFileInfoDB() (*FileInfoDatabase, error) {
	// 创建或打开文件信息数据库
	db, err := leveldb.OpenFile(fileInfoDatabasePath, nil)
	if err != nil {
		return nil, err
	}

	fileInfoDBInstance := &FileInfoDatabase{
		DB: db,
		mu: sync.Mutex{},
	}

	return fileInfoDBInstance, nil
}

// SaveFileInfo 保存文件信息
func (db *FileInfoDatabase) SaveFileInfo(ident string, info *FileInfo) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, err := info.MarshalBinary()
	if err != nil {
		return err
	}

	return db.DB.Put([]byte(ident), data, nil)
}

// GetFileInfo 获取文件信息
func (db *FileInfoDatabase) GetFileInfo(ident string) (*FileInfo, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, err := db.DB.Get([]byte(ident), nil)
	if err != nil {
		return nil, err
	}

	var info FileInfo
	if err := info.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetFileInfos 获取所有文件信息
func (db *FileInfoDatabase) GetFileInfos() (map[string]*FileInfo, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	iter := db.DB.NewIterator(nil, nil)
	defer iter.Release()

	infos := make(map[string]*FileInfo)
	for iter.Next() {
		var info FileInfo
		if err := info.UnmarshalBinary(iter.Value()); err != nil {
			return nil, err
		}
		infos[string(iter.Key())] = &info
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return infos, nil
}

// DeleteFileInfo 删除文件信息
func (db *FileInfoDatabase) DeleteFileInfo(ident string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.DB.Delete([]byte(ident), nil)
}
