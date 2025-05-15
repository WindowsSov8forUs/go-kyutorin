package processor

import (
	"fmt"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/operation"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessGroupMessage 处理群组消息
func (p *Processor) ProcessGroupMessage(payload *dto.WSPayload, data *dto.WSGroupATMessageData) error {
	// 打印消息日志
	printGroupMessage(data)

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
		Id:   data.GroupID,
		Type: channel.ChannelTypeText,
	}
	SetOpenIdType(data.GroupID, "group")

	// 构建 guild
	guild := &guild.Guild{
		Id: data.GroupID,
	}

	// 构建 member
	member := &guildmember.GuildMember{}

	// 构建 message
	message := &message.Message{
		Id:       data.ID,
		Content:  ConvertToMessageContent(data),
		CreateAt: t.UnixMilli(),
	}

	// 构建 user
	user := &user.User{
		Id:     data.Author.MemberOpenID,
		Avatar: p.getUserAvatar(data.Author.MemberOpenID),
	}

	// 填充事件数据
	event = &operation.Event{
		Sn:        id,
		Type:      operation.EventTypeMessageCreated,
		Timestamp: t.UnixMilli(),
		Login:     buildNonLoginEventLogin("qq"),
		Channel:   channel,
		Guild:     guild,
		Member:    member,
		Message:   message,
		User:      user,
	}

	// 存储消息
	messageToSave := message
	messageToSave.Channel = channel
	messageToSave.Guild = guild
	messageToSave.Member = member
	messageToSave.User = user
	database.SaveMessage(messageToSave, data.GroupID, "group")

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

func printGroupMessage(data *dto.WSGroupATMessageData) {
	// 构建消息日志
	msgContent := getMessageLog(data)

	log.Infof("收到来自群 %s 用户 %s 的消息: %s", data.GroupID, data.Author.MemberOpenID, msgContent)
}
