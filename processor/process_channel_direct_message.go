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
	"github.com/dezhishen/satori-model-go/pkg/guildmember"
	"github.com/dezhishen/satori-model-go/pkg/guildrole"
	"github.com/dezhishen/satori-model-go/pkg/message"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/websocket/client"
)

// ProcessChannelDirectMessage 处理频道私聊消息
func (p *Processor) ProcessChannelDirectMessage(data *dto.WSDirectMessageData) error {
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
		Id:   data.ChannelID,
		Type: channel.CHANNEL_TYPE_DIRECT,
	}
	// 记录私聊频道
	echo.SetDirectChannel(data.ChannelID, data.GuildID)

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
	content := handlers.ConvertToMessageContent(data)
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
		Id:        strconv.FormatInt(id, 10),
		Type:      signaling.EVENT_TYPE_MESSAGE_CREATED,
		Platform:  "qqguild",
		SelfId:    handlers.SelfId,
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
