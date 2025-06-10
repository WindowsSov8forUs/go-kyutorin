package processor

import (
	"github.com/WindowsSov8forUs/glyccat/log"
	"github.com/WindowsSov8forUs/glyccat/operation"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessGroupAddRobot 处理群组添加机器人
func (p *Processor) ProcessGroupAddRobot(payload *dto.Payload, data *dto.GroupAddBotEvent) error {
	// 输出日志
	printGroupAddRobot(data)

	// 构建事件数据
	var event *operation.Event

	// 获取事件 ID
	id := SaveEventID(payload.ID)

	// 构建 channel
	channel := &channel.Channel{
		Id:   data.GroupOpenID,
		Type: channel.ChannelTypeText,
	}
	SetOpenIdType(data.GroupOpenID, "group")

	// 构建 guild
	guild := &guild.Guild{
		Id: data.GroupOpenID,
	}

	// 构建 member
	member := &guildmember.GuildMember{}

	// 构建 user
	user := &user.User{
		Id: data.OpMemberOpenID,
	}

	// 填充事件数据
	event = &operation.Event{
		Sn:        id,
		Type:      operation.EventTypeGuildAdded,
		Timestamp: data.Timestamp,
		Login:     buildNonLoginEventLogin("qq"),
		Channel:   channel,
		Guild:     guild,
		Member:    member,
		User:      user,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

func printGroupAddRobot(data *dto.GroupAddBotEvent) {
	log.Infof("机器人被 %s 添加进了群组 %s", data.OpMemberOpenID, data.GroupOpenID)
}
