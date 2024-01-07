package processor

import (
	"fmt"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/echo"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/dezhishen/satori-model-go/pkg/message"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessC2CMessage 处理私聊消息
func (p *Processor) ProcessC2CMessage(payload *dto.WSPayload, data *dto.WSC2CMessageData) error {
	// 打印消息日志
	printC2CMessage(data)

	// 构建事件数据
	var event *signaling.Event

	// 获取事件 ID
	id, err := HashEventID(payload.ID)
	if err != nil {
		return fmt.Errorf("计算事件 ID 时出错: %v", err)
	}

	// 将事件字符串转换为时间戳
	t, err := time.Parse(time.RFC3339, string(data.Timestamp))
	if err != nil {
		return fmt.Errorf("解析时间戳时出错: %v", err)
	}

	// 构建 channel
	channel := &channel.Channel{
		Id:   data.Author.UserOpenID,
		Type: channel.CHANNEL_TYPE_DIRECT,
	}
	echo.SetOpenIdType(data.Author.UserOpenID, "private")

	// 构建 message
	message := &message.Message{
		Id:       data.ID,
		Content:  ConvertToMessageContent(data),
		CreateAt: t.Unix(),
	}

	// 构建 user
	user := &user.User{
		Id: data.Author.UserOpenID,
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        id,
		Type:      signaling.EVENT_TYPE_MESSAGE_CREATED,
		Platform:  "qq",
		SelfId:    SelfId,
		Timestamp: t.Unix(),
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

func printC2CMessage(data *dto.WSC2CMessageData) {
	// 构建消息日志
	msgContent := getMessageLog(data)

	log.Infof("收到来自用户 %s 的私聊消息: %s", data.Author.UserOpenID, msgContent)
}
