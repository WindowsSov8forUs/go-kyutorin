package processor

import (
	"fmt"
	"time"

	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/tencent-connect/botgo/dto"
)

// ProcessChannelEvent 处理频道事件
func (p *Processor) ProcessChannelEvent(payload *dto.WSPayload, data *dto.WSChannelData) error {
	// 打印消息日志
	printChannelEvent(payload, data)

	// 构建事件数据
	var event *signaling.Event

	// 获取事件 ID
	id, err := HashEventID(payload.ID)
	if err != nil {
		return fmt.Errorf("计算事件 ID 时出错: %v", err)
	}

	// 获取当前时间作为时间戳
	t := time.Now()

	// 填充事件数据
	event = &signaling.Event{
		Id:        id,
		Type:      signaling.EVENT_TYPE_INTERNAL,
		Platform:  "qqguild",
		SelfId:    SelfId,
		Timestamp: t.Unix(),
		Type_:     string(payload.Type),
		Data_:     data,
	}

	// 发送事件
	return p.BroadcastEvent(event)
}

func printChannelEvent(payload *dto.WSPayload, data *dto.WSChannelData) {
	// 构建子频道名称
	var channelName string
	channelName = fmt.Sprintf("%s ", channelTypeToString(data.Type))
	if data.Name != "" {
		channelName += fmt.Sprintf("%s(%s)", data.Name, data.ID)
	} else {
		channelName += data.ID
	}

	var logContent string

	// 根据事件类型打印不同的日志
	switch payload.Type {
	case dto.EventChannelCreate:
		logContent = fmt.Sprintf("用户 %s 在频道 %s 创建了 %s 。", data.OpUserID, data.GuildID, channelName)
	case dto.EventChannelUpdate:
		logContent = fmt.Sprintf("用户 %s 在频道 %s 更新了 %s 的信息。", data.OpUserID, data.GuildID, channelName)
	case dto.EventChannelDelete:
		logContent = fmt.Sprintf("用户 %s 在频道 %s 删除了 %s 。", data.OpUserID, data.GuildID, channelName)
	default:
		logContent = "未知的子频道事件: " + string(payload.Type)
	}

	log.Info(logContent)
}

func channelTypeToString(channelType dto.ChannelType) string {
	switch channelType {
	case dto.ChannelTypeText:
		return "文字子频道"
	case dto.ChannelTypeVoice:
		return "语音子频道"
	case dto.ChannelTypeCategory:
		return "子频道分组"
	case dto.ChannelTypeLive:
		return "直播子频道"
	case dto.ChannelTypeApplication:
		return "应用子频道"
	case dto.ChannelTypeForum:
		return "论坛子频道"
	default:
		return "未知类型子频道"
	}
}
