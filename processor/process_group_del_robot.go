package processor

import (
	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/operation"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessGroupDelRobot 处理群组删除机器人
func (p *Processor) ProcessGroupDelRobot(payload *dto.Payload, data *dto.GroupAddBotEvent) error {
	// 输出日志
	printGroupDelRobot(data)

	// 构建事件数据
	var event *operation.Event

	// 获取事件 ID
	id := SaveEventID(payload.ID)

	// 构建 channel
	channel := &channel.Channel{
		Id:   data.GroupOpenID,
		Type: channel.ChannelTypeText,
	}
	DelOpenId(data.GroupOpenID)

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
		Type:      operation.EventTypeGuildRemoved,
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

func printGroupDelRobot(data *dto.GroupAddBotEvent) {
	log.Infof("机器人被 %s 移出了群组 %s", data.OpMemberOpenID, data.GroupOpenID)
}
