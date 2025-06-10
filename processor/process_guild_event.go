package processor

import (
	"fmt"
	"time"

	"github.com/WindowsSov8forUs/glyccat/log"
	"github.com/WindowsSov8forUs/glyccat/operation"

	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessGuildEvent 处理群组事件
func (p *Processor) ProcessGuildEvent(payload *dto.Payload, data *dto.GuildData) error {
	// TODO: 有修改的可能
	var err error

	// 打印事件日志
	printGuildEvent(payload, data)

	// 构建事件数据
	var event *operation.Event

	// 获取事件 ID
	id := SaveEventID(payload.ID)

	// 根据不同的 payload.Type 设置不同的 event.Type
	var eventType operation.EventType
	switch payload.Type {
	case dto.EventGuildCreate:
		eventType = operation.EventTypeGuildAdded
	case dto.EventGuildUpdate:
		eventType = operation.EventTypeGuildUpdated
	case dto.EventGuildDelete:
		eventType = operation.EventTypeGuildRemoved
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
		Id:     data.ID,
		Name:   data.Name,
		Avatar: data.Icon,
	}

	// 构建 operator
	operator := &user.User{
		Id: data.OpUserID,
	}

	// 填充事件数据
	event = &operation.Event{
		Sn:        id,
		Type:      eventType,
		Timestamp: t.UnixMilli(),
		Login:     buildNonLoginEventLogin("qqguild"),
		Guild:     guild,
		Operator:  operator,
	}

	// 发送事件
	return p.BroadcastEvent(event)
}

func printGuildEvent(payload *dto.Payload, data *dto.GuildData) {
	// 构建频道名称
	var guildName string
	if data.Name != "" {
		guildName = fmt.Sprintf("%s(%s)", data.Name, data.ID)
	} else {
		guildName = data.ID
	}

	// 构建日志内容
	var logContent string
	switch payload.Type {
	case dto.EventGuildCreate:
		logContent = fmt.Sprintf("用户 %s 创建了频道 %s 。", data.OpUserID, guildName)
	case dto.EventGuildUpdate:
		logContent = fmt.Sprintf("用户 %s 更新了频道 %s 的信息。", data.OpUserID, guildName)
	case dto.EventGuildDelete:
		logContent = fmt.Sprintf("用户 %s 删除了频道 %s 。", data.OpUserID, guildName)
	default:
		logContent = "未知的频道事件: " + string(payload.Type)
	}

	// 打印日志
	log.Info(logContent)
}
