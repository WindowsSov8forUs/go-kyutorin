package fileserver

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const metaDatabasePath = "data/files/.metadata"

// FileMetadata 文件元数据
type FileMetadata struct {
	ID           string      `json:"id"`           // 文件唯一标识 ID
	Name         string      `json:"name"`         // 文件名
	URL          string      `json:"url"`          // 文件内部链接
	Path         string      `json:"path"`         // 文件存储相对路径
	ContentType  string      `json:"content_type"` // 文件内容类型
	CreateAt     uint64      `json:"create_at"`    // 文件创建时间戳
	TTL          uint64      `json:"ttl"`          // 文件有效时间
	cleanerTimer *time.Timer `json:"-"`            // 清理定时器，用于定时删除过期文件
}

// MarshalBinary 序列化文件元数据为二进制
func (m *FileMetadata) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary 反序列化二进制数据为文件元数据
func (m *FileMetadata) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	if err := json.NewDecoder(buf).Decode(m); err != nil {
		return err
	}
	return nil
}

// MetaDatabase 文件元数据数据库
type MetaDatabase struct {
	DB *leveldb.DB
	mu sync.Mutex
}

// StartMetaDB 启动文件元数据数据库
func StartMetaDB() (*MetaDatabase, error) {
	// 创建或打开元数据数据库
	db, err := leveldb.OpenFile(metaDatabasePath, nil)
	if err != nil {
		return nil, err
	}

	metaDBInstance := &MetaDatabase{
		DB: db,
		mu: sync.Mutex{},
	}

	return metaDBInstance, nil
}

// SaveFileMeta 保存文件元数据
func (db *MetaDatabase) SaveFileMeta(ident string, meta *FileMetadata) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, err := meta.MarshalBinary()
	if err != nil {
		return err
	}

	return db.DB.Put([]byte(ident), data, nil)
}

// GetFileMeta 获取文件元数据
func (db *MetaDatabase) GetFileMeta(ident string) (*FileMetadata, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	data, err := db.DB.Get([]byte(ident), nil)
	if err != nil {
		return nil, err
	}

	var meta FileMetadata
	if err := meta.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	return &meta, nil
}

// GetFileMetas 获取所有文件元数据
func (db *MetaDatabase) GetFileMetas() (map[string]*FileMetadata, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	iter := db.DB.NewIterator(nil, nil)
	defer iter.Release()

	metas := make(map[string]*FileMetadata)
	for iter.Next() {
		var meta FileMetadata
		if err := meta.UnmarshalBinary(iter.Value()); err != nil {
			return nil, err
		}
		metas[string(iter.Key())] = &meta
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return metas, nil
}

// DeleteFileMeta 删除文件元数据
func (db *MetaDatabase) DeleteFileMeta(ident string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.DB.Delete([]byte(ident), nil)
}
