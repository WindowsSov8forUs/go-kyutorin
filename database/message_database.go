package database

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
	"sync"

	"github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const messageDBPath string = "data/cache/message"

// MessageDB 消息数据库
type MessageDB struct {
	DB    *leveldb.DB
	mu    sync.Mutex
	limit int
}

var messageDBInstance *MessageDB

// StartMessageDB 启动消息数据库
func StartMessageDB(messageLimit int) error {
	// 创建或打开消息缓存数据库
	db, err := leveldb.OpenFile(messageDBPath, nil)
	if err != nil {
		return err
	}

	messageDBInstance = &MessageDB{
		DB:    db,
		limit: messageLimit,
	}

	return nil
}

// SaveMessage 保存消息
func SaveMessage(data *message.Message, channelId, channelType string) error {
	messageDBInstance.mu.Lock()
	defer messageDBInstance.mu.Unlock()

	// 保存消息
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return err
	}

	// 保存消息
	key := fmt.Sprintf("%s:%s:%s", channelType, channelId, data.Id)
	if err := messageDBInstance.DB.Put([]byte(key), buf.Bytes(), nil); err != nil {
		return err
	}

	return nil
}

// GetMessage 获取消息
func GetMessage(channelId, channelType, messageId string) (*message.Message, error) {
	messageDBInstance.mu.Lock()
	defer messageDBInstance.mu.Unlock()

	// 获取消息
	key := fmt.Sprintf("%s:%s:%s", channelType, channelId, messageId)
	data, err := messageDBInstance.DB.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}

	// 解码消息
	var message message.Message
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&message); err != nil {
		return nil, err
	}

	return &message, nil
}

// GetMessageList 获取消息列表
func GetMessageList(channelId, channelType, next string) ([]*message.Message, string, error) {
	messageDBInstance.mu.Lock()
	defer messageDBInstance.mu.Unlock()

	// 获取消息列表
	prefix := []byte(fmt.Sprintf("%s:%s", channelType, channelId))
	iter := messageDBInstance.DB.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()

	var messageList []*message.Message
	var nextMessage *message.Message

	if next != "" {
		// 查找 next 指定的消息
		nextKey := []byte(fmt.Sprintf("%s:%s:%s", channelType, channelId, next))
		nextValue, err := messageDBInstance.DB.Get(nextKey, nil)
		if err != nil {
			return nil, "", err
		}
		dec := gob.NewDecoder(bytes.NewReader(nextValue))
		if err := dec.Decode(&nextMessage); err != nil {
			return nil, "", err
		}
	}

	for iter.Next() {
		// 解码消息
		var message message.Message
		dec := gob.NewDecoder(bytes.NewReader(iter.Value()))
		if err := dec.Decode(&message); err != nil {
			continue
		}

		// 将 CreateAt 大于 next 的消息添加到 messageList
		if nextMessage == nil || message.CreateAt > nextMessage.CreateAt {
			messageList = append(messageList, &message)
		}
	}

	// 按 CreateAt 对 messageList 进行排序
	sort.Slice(messageList, func(i, j int) bool {
		return messageList[i].CreateAt < messageList[j].CreateAt
	})

	next = "" // 重置 next

	// 如果消息列表已满，则设置 next 为最后一条消息的 Id
	if messageDBInstance.limit > 0 && len(messageList) > messageDBInstance.limit {
		messageList = messageList[:messageDBInstance.limit]
		next = messageList[messageDBInstance.limit-1].Id
	}

	return messageList, next, nil
}
