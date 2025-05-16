package processor

import (
	"fmt"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/operation"

	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessMemberEvent 处理群组成员事件
func (p *Processor) ProcessMemberEvent(payload *dto.Payload, data *dto.GuildMemberData) error {
	// TODO: 有修改的可能
	var err error

	// 打印事件日志
	printMemberEvent(payload, data)

	// 构建事件数据
	var event *operation.Event

	// 获取事件 ID
	id := SaveEventID(payload.ID)

	// 根据不同的 payload.Type 设置不同的 event.Type
	var eventType operation.EventType
	switch payload.Type {
	case dto.EventGuildMemberAdd:
		eventType = operation.EventTypeGuildMemberAdded
	case dto.EventGuildMemberUpdate:
		eventType = operation.EventTypeGuildMemberUpdated
	case dto.EventGuildMemberRemove:
		eventType = operation.EventTypeGuildMemberRemoved
	default:
		return fmt.Errorf("未知的 payload.Type: %v", payload.Type)
	}

	// 根据不同的 payload.Type 通过不同方式获取 Timestamp
	var t time.Time
	if payload.Type == dto.EventGuildCreate {
		t, err = time.Parse(time.RFC3339, string(data.JoinedAt))
		if err != nil {
			return fmt.Errorf("解析时间戳时出错: %v", err)
		}
	} else {
		// 获取当前时间作为时间戳
		t = time.Now()
	}

	// 构建 guild
	guild := &guild.Guild{
		Id: data.GuildID,
	}

	// 构建 member
	member := &guildmember.GuildMember{
		Nick: data.Nick,
	}
	// 获取加入时间
	joinedAt, err := data.JoinedAt.Time()
	if err != nil {
		return fmt.Errorf("解析时间戳时出错: %v", err)
	}
	member.JoinedAt = joinedAt.UnixMilli()

	// 构建 operator
	operator := &user.User{
		Id: data.OpUserID,
	}

	// 构建 user
	user := &user.User{
		Id:     data.User.ID,
		Name:   data.User.Username,
		Avatar: data.User.Avatar,
		IsBot:  data.User.Bot,
	}

	// 填充事件数据
	event = &operation.Event{
		Sn:        id,
		Type:      eventType,
		Timestamp: t.UnixMilli(),
		Login:     buildNonLoginEventLogin("qqguild"),
		Guild:     guild,
		Member:    member,
		Operator:  operator,
		User:      user,
	}

	// 发送事件
	return p.BroadcastEvent(event)
}

func printMemberEvent(payload *dto.Payload, data *dto.GuildMemberData) {
	// 构建成员名称
	var memberName string
	if data.Nick != "" {
		memberName = fmt.Sprintf("%s(%s)", data.Nick, data.User.ID)
	} else if data.User.Username != "" {
		memberName = fmt.Sprintf("%s(%s)", data.User.Username, data.User.ID)
	} else {
		memberName = data.User.ID
	}

	// 构建日志内容
	var logContent string
	switch payload.Type {
	case dto.EventGuildMemberAdd:
		if data.User.ID == data.OpUserID {
			logContent = fmt.Sprintf("用户 %s 加入了频道 %s 。", memberName, data.GuildID)
		} else {
			logContent = fmt.Sprintf("用户 %s 邀请了用户 %s 加入频道 %s 。", data.OpUserID, memberName, data.GuildID)
		}
	case dto.EventGuildMemberUpdate:
		if data.User.ID == data.OpUserID {
			logContent = fmt.Sprintf("频道 %s 的用户 %s 更新了自己的信息。", data.GuildID, memberName)
		} else {
			logContent = fmt.Sprintf("频道 %s 的用户 %s 更新了用户 %s 的信息。", data.GuildID, data.OpUserID, memberName)
		}
	case dto.EventGuildMemberRemove:
		if data.User.ID == data.OpUserID {
			logContent = fmt.Sprintf("用户 %s 离开了频道 %s 。", memberName, data.GuildID)
		} else {
			logContent = fmt.Sprintf("用户 %s 将用户 %s 移出了频道 %s 。", data.OpUserID, memberName, data.GuildID)
		}
	default:
		logContent = "未知的频道成员事件: " + string(payload.Type)
	}

	// 打印日志
	log.Info(logContent)
}
