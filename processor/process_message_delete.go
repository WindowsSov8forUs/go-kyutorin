package processor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/handlers"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/dezhishen/satori-model-go/pkg/message"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessMessageDelete 将消息撤回事件转换为 Satori 的 MessageDeleted 事件
func (p *Processor) ProcessMessageDelete(payload *dto.WSPayload, data interface{}) error {
	// 强制类型转换获取 MessageDelete 结构
	var messageDelete *dto.MessageDelete
	var channelType channel.ChannelType // 获取平台名称
	switch v := data.(type) {
	case *dto.WSMessageDeleteData:
		messageDelete = (*dto.MessageDelete)(v)
		channelType = channel.CHANNEL_TYPE_TEXT
	case *dto.WSPublicMessageDeleteData:
		messageDelete = (*dto.MessageDelete)(v)
		channelType = channel.CHANNEL_TYPE_TEXT
	case *dto.WSDirectMessageDeleteData:
		messageDelete = (*dto.MessageDelete)(v)
		channelType = channel.CHANNEL_TYPE_DIRECT
	default:
		return fmt.Errorf("无法处理的消息撤回事件: %v", data)
	}

	// 构建事件数据
	var event *signaling.Event

	// 获取 s
	id := strconv.FormatInt(payload.S, 10)

	// 将当前时间转换为时间戳
	t := time.Now().Unix()

	// 构建 channel
	channel := &channel.Channel{
		Id:   messageDelete.Message.ChannelID,
		Type: channelType,
	}

	// 构建 guild
	guild := &guild.Guild{
		Id: messageDelete.Message.GuildID,
	}

	// 构建 message
	message := &message.Message{
		Id: messageDelete.Message.ID,
	}

	// 构建 operator
	operator := &user.User{
		Id:     messageDelete.OpUser.ID,
		Name:   messageDelete.OpUser.Username,
		Avatar: messageDelete.OpUser.Avatar,
		IsBot:  messageDelete.OpUser.Bot,
	}

	// 构建 user
	user := &user.User{
		Id:     messageDelete.Message.Author.ID,
		Name:   messageDelete.Message.Author.Username,
		Avatar: messageDelete.Message.Author.Avatar,
		IsBot:  messageDelete.Message.Author.Bot,
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        id,
		Type:      signaling.EVENT_TYPE_MESSAGE_DELETED,
		Platform:  "qqguild",
		SelfId:    handlers.SelfId,
		Timestamp: t,
		Channel:   channel,
		Guild:     guild,
		Message:   message,
		Operator:  operator,
		User:      user,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}
