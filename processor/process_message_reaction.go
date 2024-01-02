package processor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/echo"
	"github.com/WindowsSov8forUs/go-kyutorin/handlers"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/dezhishen/satori-model-go/pkg/message"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessMessageReaction 处理消息回应
func (p *Processor) ProcessMessageReaction(payload *dto.WSPayload, data *dto.WSMessageReactionData) error {
	// TODO: 更好的处理方式

	// 构建事件数据
	var event *signaling.Event

	// 获取 s
	id := strconv.FormatInt(payload.S, 10)

	// 根据 payload.Type 判断事件类型
	var eventType signaling.EventType
	switch payload.Type {
	case dto.EventMessageReactionAdd:
		eventType = signaling.EVENT_TYPE_REACTION_ADDED
	case dto.EventMessageReactionRemove:
		eventType = signaling.EVENT_TYPE_REACTION_REMOVED
	default:
		return fmt.Errorf("无法处理的消息回应事件: %v", data)
	}

	// 将当前时间转换为时间戳
	t := time.Now().Unix()

	// 获取频道类型
	var channelType channel.ChannelType
	guildId := echo.GetDirectChannelGuild(data.ChannelID)
	if guildId != "" {
		channelType = channel.CHANNEL_TYPE_DIRECT
	} else {
		channelType = channel.CHANNEL_TYPE_TEXT
	}
	// 构建 channel
	channel := &channel.Channel{
		Id:   data.ChannelID,
		Type: channelType,
	}

	// 构建 guild
	guild := &guild.Guild{
		Id: data.GuildID,
	}

	// 构建 message
	var m *message.Message
	// 根据 Target.Type 判断消息类型
	if data.Target.Type == 0 {
		// 是消息
		m = &message.Message{
			Id: data.Target.ID,
		}
	} else {
		// 不赋值
		m = nil
	}

	// 构建 operator
	operator := &user.User{
		Id: data.UserID,
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        id,
		Type:      eventType,
		Platform:  "qqguild",
		SelfId:    handlers.SelfId,
		Timestamp: t,
		Channel:   channel,
		Guild:     guild,
		Message:   m,
		Operator:  operator,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}
