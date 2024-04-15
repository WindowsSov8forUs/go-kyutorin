package processor

import (
	"fmt"
	"time"

	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildrole"
	"github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessGuildNormalMessage 处理群组私域消息
func (p *Processor) ProcessGuildNormalMessage(payload *dto.WSPayload, data *dto.WSMessageData) error {
	// 打印消息日志
	printGuildMessage(data)

	// 构建事件数据
	var event *signaling.Event

	// 获取事件 ID
	id := RecordEventID(payload.ID)

	// 将事件字符串转换为时间戳
	t, err := time.Parse(time.RFC3339, string(data.Timestamp))
	if err != nil {
		return fmt.Errorf("解析时间戳时出错: %v", err)
	}

	// 构建 channel
	channel := &channel.Channel{
		Id:   data.ChannelID,
		Type: channel.CHANNEL_TYPE_TEXT,
	}

	// 构建 guild
	guild := &guild.Guild{
		Id: data.GuildID,
	}

	// 构建 member
	joinedTime, err := data.Member.JoinedAt.Time()
	if err != nil {
		return fmt.Errorf("解析加入时间时出错: %v", err)
	}
	member := &guildmember.GuildMember{
		Nick:     data.Member.Nick,
		JoinedAt: joinedTime.Unix(),
	}

	// 构建 message
	message := &message.Message{
		Id:       data.ID,
		CreateAt: t.Unix(),
	}
	// 转换消息格式
	content := ConvertToMessageContent(data)
	message.Content = content

	// 构建 user
	user := &user.User{
		Id:     data.Author.ID,
		Name:   data.Author.Username,
		Avatar: data.Author.Avatar,
		IsBot:  data.Author.Bot,
	}

	// 构建 role
	role := &guildrole.GuildRole{
		Id: data.Member.Roles[0],
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        id,
		Type:      signaling.EventTypeMessageCreated,
		Platform:  "qqguild",
		SelfId:    SelfId,
		Timestamp: t.Unix(),
		Channel:   channel,
		Guild:     guild,
		Member:    member,
		Message:   message,
		Role:      role,
		User:      user,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

func printGuildMessage(data *dto.WSMessageData) {
	// 构建用户名称
	var userName string
	if data.Member.Nick != "" {
		userName = fmt.Sprintf("%s(%s)", data.Member.Nick, data.Author.ID)
	} else if data.Author.Username != "" {
		userName = fmt.Sprintf("%s(%s)", data.Author.Username, data.Author.ID)
	} else {
		userName = data.Author.ID
	}

	// 构建消息日志
	msgContent := getMessageLog(data)

	log.Infof("收到来自频道 %s 的子频道 %s 的用户 %s 的消息: %s", data.GuildID, data.ChannelID, userName, msgContent)
}
