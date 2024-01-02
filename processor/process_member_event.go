package processor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/handlers"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/dezhishen/satori-model-go/pkg/guildmember"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessMemberEvent 处理群组成员事件
func (p *Processor) ProcessMemberEvent(payload *dto.WSPayload, data *dto.WSGuildMemberData) error {
	// TODO: 有修改的可能

	// 构建事件数据
	var event *signaling.Event

	// 获取 s
	id := payload.S

	// 根据不同的 payload.Type 设置不同的 event.Type
	var eventType signaling.EventType
	switch payload.Type {
	case dto.EventGuildMemberAdd:
		eventType = signaling.EVENT_TYPE_GUILD_MEMBER_ADDED
	case dto.EventGuildMemberUpdate:
		eventType = signaling.EVENT_TYPE_GUILD_MEMBER_UPDATED
	case dto.EventGuildMemberRemove:
		eventType = signaling.EVENT_TYPE_GUILD_MEMBER_REMOVED
	default:
		return fmt.Errorf("未知的 payload.Type: %v", payload.Type)
	}

	// 根据不同的 payload.Type 通过不同方式获取 Timestamp
	var t time.Time
	var err error
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
	member.JoinedAt = joinedAt.Unix()

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
	event = &signaling.Event{
		Id:        strconv.FormatInt(id, 10),
		Type:      eventType,
		Platform:  "qqguild",
		SelfId:    handlers.SelfId,
		Timestamp: t.Unix(),
		Guild:     guild,
		Member:    member,
		Operator:  operator,
		User:      user,
	}

	// 发送事件
	return p.BroadcastEvent(event)
}
