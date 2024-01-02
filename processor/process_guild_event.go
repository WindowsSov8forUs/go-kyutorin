package processor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/handlers"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessGuildEvent 处理群组事件
func (p *Processor) ProcessGuildEvent(payload *dto.WSPayload, data *dto.WSGuildData) error {
	// TODO: 有修改的可能

	// 构建事件数据
	var event *signaling.Event

	// 获取 s
	id := payload.S

	// 根据不同的 payload.Type 设置不同的 event.Type
	var eventType signaling.EventType
	switch payload.Type {
	case dto.EventGuildCreate:
		eventType = signaling.EVENT_TYPE_GUILD_ADDED
	case dto.EventGuildUpdate:
		eventType = signaling.EVENT_TYPE_GUILD_UPDATED
	case dto.EventGuildDelete:
		eventType = signaling.EVENT_TYPE_GUILD_REMOVED
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
		Id:     data.ID,
		Name:   data.Name,
		Avatar: data.Icon,
	}

	// 构建 operator
	operator := &user.User{
		Id: data.OpUserID,
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        strconv.FormatInt(id, 10),
		Type:      eventType,
		Platform:  "qqguild",
		SelfId:    handlers.SelfId,
		Timestamp: t.Unix(),
		Guild:     guild,
		Operator:  operator,
	}

	// 发送事件
	return p.BroadcastEvent(event)
}
