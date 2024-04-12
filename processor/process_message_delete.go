package processor

import (
	"fmt"
	"time"

	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
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

	// 打印消息日志
	printMessageDeleteEvent(payload, messageDelete)

	// 构建事件数据
	var event *signaling.Event

	// 获取事件 ID
	id := RecordEventID(payload.ID)

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
		SelfId:    SelfId,
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

func printMessageDeleteEvent(payload *dto.WSPayload, data *dto.MessageDelete) {
	// 构建用户名称
	var userName string
	if data.OpUser.Username != "" {
		userName = fmt.Sprintf("%s(%s)", data.OpUser.Username, data.OpUser.ID)
	} else {
		userName = data.OpUser.ID
	}

	// 构建成员名称
	var memberName string
	if data.Message.Member.Nick != "" {
		memberName = fmt.Sprintf("%s(%s)", data.Message.Member.Nick, data.Message.Author.ID)
	} else if data.Message.Author.Username != "" {
		memberName = fmt.Sprintf("%s(%s)", data.Message.Author.Username, data.Message.Author.ID)
	} else {
		memberName = data.Message.Author.ID
	}

	// 构建日志内容
	var logContent string
	switch payload.Type {
	case dto.EventMessageDelete:
		if data.OpUser.ID == data.Message.Author.ID {
			logContent = fmt.Sprintf("频道 %s 的子频道 %s 的用户 %s 撤回了一条消息。", data.Message.GuildID, data.Message.ChannelID, userName)
		} else {
			logContent = fmt.Sprintf("频道 %s 的子频道 %s 的用户 %s 撤回了用户 %s 的一条消息。", data.Message.GuildID, data.Message.ChannelID, userName, memberName)
		}
	case dto.EventPublicMessageDelete:
		if data.OpUser.ID == data.Message.Author.ID {
			logContent = fmt.Sprintf("频道 %s 的子频道 %s 的用户 %s 撤回了一条消息。", data.Message.GuildID, data.Message.ChannelID, userName)
		} else {
			logContent = fmt.Sprintf("频道 %s 的子频道 %s 的用户 %s 撤回了用户 %s 的一条消息。", data.Message.GuildID, data.Message.ChannelID, userName, memberName)
		}
	case dto.EventDirectMessageDelete:
		logContent = fmt.Sprintf("用户 %s 撤回了一条私聊频道消息。", userName)
	default:
		if data.OpUser.ID == data.Message.Author.ID {
			logContent = fmt.Sprintf("用户 %s 撤回了一条消息。", userName)
		} else {
			logContent = fmt.Sprintf("用户 %s 撤回了用户 %s 的一条消息。", userName, memberName)
		}
	}

	// 打印日志
	log.Info(logContent)
}
