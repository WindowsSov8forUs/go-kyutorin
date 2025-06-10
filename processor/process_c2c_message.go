package processor

import (
	"fmt"
	"time"

	"github.com/WindowsSov8forUs/glyccat/database"
	"github.com/WindowsSov8forUs/glyccat/log"
	"github.com/WindowsSov8forUs/glyccat/operation"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessC2CMessage 处理私聊消息
func (p *Processor) ProcessC2CMessage(payload *dto.Payload, data *dto.C2CMessageData) error {
	// 打印消息日志
	printC2CMessage(data)

	// 构建事件数据
	var event *operation.Event

	// 获取事件 ID
	id := SaveEventID(payload.ID)

	// 将事件字符串转换为时间戳
	t, err := time.Parse(time.RFC3339, string(data.Timestamp))
	if err != nil {
		return fmt.Errorf("解析时间戳时出错: %v", err)
	}

	// 构建 channel
	channel := &channel.Channel{
		Id:   data.Author.UserOpenID,
		Type: channel.ChannelTypeDirect,
	}
	SetOpenIdType(data.Author.UserOpenID, "private")

	// 构建 message
	message := &message.Message{
		Id:       data.ID,
		Content:  ConvertToMessageContent(data),
		CreateAt: t.UnixMilli(),
	}

	// 构建 user
	user := &user.User{
		Id:     data.Author.UserOpenID,
		Avatar: p.getUserAvatar(data.Author.UserOpenID),
	}

	// 填充事件数据
	event = &operation.Event{
		Sn:        id,
		Type:      operation.EventTypeMessageCreated,
		Timestamp: t.UnixMilli(),
		Login:     buildNonLoginEventLogin("qq"),
		Channel:   channel,
		Message:   message,
		User:      user,
	}

	// 存储消息
	messageToSave := message
	messageToSave.Channel = channel
	messageToSave.User = user
	database.SaveMessage(messageToSave, data.Author.UserOpenID, "private")

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

func printC2CMessage(data *dto.C2CMessageData) {
	// 构建消息日志
	msgContent := getMessageLog(data)

	log.Infof("收到来自用户 %s 的私聊消息: %s", data.Author.UserOpenID, msgContent)
}
