package processor

import (
	"github.com/WindowsSov8forUs/go-kyutorin/echo"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessGroupDelRobot 处理群组删除机器人
func (p *Processor) ProcessGroupDelRobot(payload *dto.WSPayload, data *dto.WSGroupDelRobotData) error {
	// 输出日志
	printGroupDelRobot(data)

	// 构建事件数据
	var event *signaling.Event

	// 获取事件 ID
	id := RecordEventID(payload.ID)

	// 构建 channel
	channel := &channel.Channel{
		Id:   data.GroupOpenid,
		Type: channel.CHANNEL_TYPE_TEXT,
	}
	echo.DelOpenId(data.GroupOpenid)

	// 构建 guild
	guild := &guild.Guild{
		Id: data.GroupOpenid,
	}

	// 构建 member
	member := &guildmember.GuildMember{}

	// 构建 user
	user := &user.User{
		Id: data.OpMemberOpenid,
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        id,
		Type:      signaling.EventTypeGuildRemoved,
		Platform:  "qq",
		SelfId:    SelfId,
		Timestamp: data.Timestamp,
		Channel:   channel,
		Guild:     guild,
		Member:    member,
		User:      user,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

func printGroupDelRobot(data *dto.WSGroupDelRobotData) {
	log.Infof("机器人被 %s 移出了群组 %s", data.OpMemberOpenid, data.GroupOpenid)
}
