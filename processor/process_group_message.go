package processor

import (
	"fmt"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/echo"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/dezhishen/satori-model-go/pkg/guildmember"
	"github.com/dezhishen/satori-model-go/pkg/message"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessGroupMessage 处理群组消息
func (p *Processor) ProcessGroupMessage(payload *dto.WSPayload, data *dto.WSGroupATMessageData) error {
	// 打印消息日志
	printGroupMessage(data)

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
		Id:   data.GroupID,
		Type: channel.CHANNEL_TYPE_TEXT,
	}
	echo.SetOpenIdType(data.GroupID, "group")

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
		CreateAt: t.Unix(),
	}

	// 构建 user
	user := &user.User{
		Id: data.Author.MemberOpenID,
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        id,
		Type:      signaling.EVENT_TYPE_MESSAGE_CREATED,
		Platform:  "qq",
		SelfId:    SelfId,
		Timestamp: t.Unix(),
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
