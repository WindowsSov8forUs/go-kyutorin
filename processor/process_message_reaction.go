package processor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/echo"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessMessageReaction 处理消息回应
func (p *Processor) ProcessMessageReaction(payload *dto.WSPayload, data *dto.WSMessageReactionData) error {
	// TODO: 更好的处理方式

	// 打印消息日志
	printMessageReaction(payload, data)

	// 构建事件数据
	var event *signaling.Event

	// 获取事件 ID
	id := RecordEventID(payload.ID)

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
		SelfId:    SelfId,
		Timestamp: t,
		Channel:   channel,
		Guild:     guild,
		Message:   m,
		Operator:  operator,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

func printMessageReaction(payload *dto.WSPayload, data *dto.WSMessageReactionData) {
	// 构建目标名称
	targetName := fmt.Sprintf("%s(%s)", targetTypeToString(data.Target.Type), data.Target.ID)

	// 构建 Emoji 名称
	emojiName := fmt.Sprintf("%s(%s)", emojiTypeToString(data.Emoji.Type), data.Emoji.ID)

	var logContent string
	switch payload.Type {
	case dto.EventMessageReactionAdd:
		logContent = fmt.Sprintf("频道 %s 的子频道 %s 的用户 %s 对 %s 进行了表态: %s", data.GuildID, data.ChannelID, data.UserID, targetName, emojiName)
	case dto.EventMessageReactionRemove:
		logContent = fmt.Sprintf("频道 %s 的子频道 %s 的用户 %s 对 %s 移除了表态: %s", data.GuildID, data.ChannelID, data.UserID, targetName, emojiName)
	default:
		logContent = fmt.Sprintf("频道 %s 的子频道 %s 的用户 %s 对 %s 发生了表态事件: %s", data.GuildID, data.ChannelID, data.UserID, targetName, emojiName)
	}

	log.Info(logContent)
}

func targetTypeToString(targetType dto.ReactionTargetType) string {
	switch targetType {
	case dto.ReactionTargetTypeMsg:
		return "消息"
	case dto.ReactionTargetTypeFeed:
		return "帖子"
	case dto.ReactionTargetTypeComment:
		return "评论"
	case dto.ReactionTargetTypeReply:
		return "回复"
	default:
		return "[" + strconv.Itoa(int(targetType)) + "]"
	}
}

func emojiTypeToString(emojiType int) string {
	switch emojiType {
	case 1:
		return "系统表情"
	case 2:
		return "emoji表情"
	default:
		return "[表情" + strconv.Itoa(int(emojiType)) + "]"
	}
}
