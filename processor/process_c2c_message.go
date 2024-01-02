package processor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/handlers"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/dezhishen/satori-model-go/pkg/message"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/websocket/client"
)

// ProcessC2CMessage 处理私聊消息
func (p *Processor) ProcessC2CMessage(data *dto.WSC2CMessageData) error {
	// 构建事件数据
	var event *signaling.Event

	// 获取 s
	id := client.GetGlobalS()

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

	// 构建 message
	message := &message.Message{
		Id:       data.ID,
		Content:  handlers.ConvertToMessageContent(data),
		CreateAt: t.Unix(),
	}

	// 构建 user
	user := &user.User{
		Id: data.Author.UserOpenID,
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        strconv.FormatInt(id, 10),
		Type:      signaling.EVENT_TYPE_MESSAGE_CREATED,
		Platform:  "qq",
		SelfId:    handlers.SelfId,
		Timestamp: t.Unix(),
		Channel:   channel,
		Message:   message,
		User:      user,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}
