package event

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/tidwall/gjson" // 由于回包的 d 类型不确定，gjson 用于从回包json中提取 d 并进行针对性的解析

	"github.com/tencent-connect/botgo/dto"
)

func init() {
	// Start a goroutine for periodic cleaning
	go cleanProcessedIDs()
}

func cleanProcessedIDs() {
	ticker := time.NewTicker(5 * time.Minute) // Adjust the interval as needed
	defer ticker.Stop()

	for range ticker.C {
		// Clean processedIDs, remove entries which are no longer needed
		processedIDs.Range(func(key, value interface{}) bool {
			processedIDs.Delete(key)
			return true
		})
	}
}

var processedIDs sync.Map

var eventParseFuncMap = map[dto.OPCode]map[dto.EventType]eventParseFunc{
	dto.DispatchEvent: {
		dto.EventGuildCreate: guildHandler,
		dto.EventGuildUpdate: guildHandler,
		dto.EventGuildDelete: guildHandler,

		dto.EventChannelCreate: channelHandler,
		dto.EventChannelUpdate: channelHandler,
		dto.EventChannelDelete: channelHandler,

		dto.EventGuildMemberAdd:    guildMemberHandler,
		dto.EventGuildMemberUpdate: guildMemberHandler,
		dto.EventGuildMemberRemove: guildMemberHandler,

		dto.EventMessageCreate: messageHandler,
		dto.EventMessageDelete: messageDeleteHandler,

		dto.EventMessageReactionAdd:    messageReactionHandler,
		dto.EventMessageReactionRemove: messageReactionHandler,

		dto.EventAtMessageCreate:     atMessageHandler,
		dto.EventPublicMessageDelete: publicMessageDeleteHandler,

		dto.EventDirectMessageCreate: directMessageHandler,
		dto.EventDirectMessageDelete: directMessageDeleteHandler,

		dto.EventAudioStart:  audioHandler,
		dto.EventAudioFinish: audioHandler,
		dto.EventAudioOnMic:  audioHandler,
		dto.EventAudioOffMic: audioHandler,

		dto.EventMessageAuditPass:   messageAuditHandler,
		dto.EventMessageAuditReject: messageAuditHandler,

		dto.EventForumThreadCreate: threadHandler,
		dto.EventForumThreadUpdate: threadHandler,
		dto.EventForumThreadDelete: threadHandler,
		dto.EventForumPostCreate:   postHandler,
		dto.EventForumPostDelete:   postHandler,
		dto.EventForumReplyCreate:  replyHandler,
		dto.EventForumReplyDelete:  replyHandler,
		dto.EventForumAuditResult:  forumAuditHandler,

		dto.EventInteractionCreate:    interactionHandler,
		dto.EventGroupAtMessageCreate: groupAtMessageHandler,
		dto.EventC2CMessageCreate:     c2cMessageHandler,
		dto.EventGroupAddRobot:        groupaddbothandler,
		dto.EventGroupDelRobot:        groupdelbothandler,
		dto.EventGroupMsgReject:       groupMsgRejecthandler,
		dto.EventGroupMsgReceive:      groupMsgReceivehandler,
	},
}

type eventParseFunc func(event *dto.Payload, message []byte) error

// ParseAndHandle 处理回调事件
func ParseAndHandle(payload *dto.Payload) error {
	// 指定类型的 handler
	if h, ok := eventParseFuncMap[payload.OPCode][payload.Type]; ok {
		return h(payload, payload.RawMessage)
	}
	// 透传handler，如果未注册具体类型的 handler，会统一投递到这个 handler
	if DefaultHandlers.Plain != nil {
		return DefaultHandlers.Plain(payload, payload.RawMessage)
	}
	return nil
}

// ParseData 解析数据
func ParseData(message []byte, target interface{}) error {
	// 获取数据部分
	data := gjson.Get(string(message), "d")
	// 外层ID 与内层ID不同 外层id是event_id 用于发送参数 d内层id是id,用于put回调接口
	eventid := gjson.Get(string(message), "id").String()

	// 使用switch语句处理不同类型
	switch v := target.(type) {
	case *dto.ThreadData:
		// 特殊处理dto.ThreadData
		if err := json.Unmarshal([]byte(data.String()), v); err != nil {
			return err
		}
		// 设置ID字段
		v.EventID = eventid
		return nil

	case *dto.GroupAddBotEvent:
		// 特殊处理dto.GroupAddBotEvent
		if err := json.Unmarshal([]byte(data.String()), v); err != nil {
			return err
		}
		// 设置ID字段
		v.EventID = eventid
		return nil

	case *dto.InteractionEventData:
		// 特殊处理dto.InteractionEventData
		if err := json.Unmarshal([]byte(data.String()), v); err != nil {
			return err
		}
		// 设置ID字段
		v.EventID = eventid
		return nil

	case *dto.GroupMsgRejectEvent:
		// 特殊处理dto.GroupMsgRejectEvent
		if err := json.Unmarshal([]byte(data.String()), v); err != nil {
			return err
		}
		// 设置ID字段
		v.EventID = eventid
		return nil

	case *dto.GroupMsgReceiveEvent:
		// 特殊处理dto.GroupMsgReceiveEvent
		if err := json.Unmarshal([]byte(data.String()), v); err != nil {
			return err
		}
		// 设置ID字段
		v.EventID = eventid
		return nil

	default:
		// 对于其他类型，继续原有逻辑
		return json.Unmarshal([]byte(data.String()), target)
	}
}

func guildHandler(payload *dto.Payload, message []byte) error {
	data := &dto.GuildData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.Guild != nil {
		return DefaultHandlers.Guild(payload, data)
	}
	return nil
}

func channelHandler(payload *dto.Payload, message []byte) error {
	data := &dto.ChannelData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.Channel != nil {
		return DefaultHandlers.Channel(payload, data)
	}
	return nil
}

func guildMemberHandler(payload *dto.Payload, message []byte) error {
	data := &dto.GuildMemberData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.GuildMember != nil {
		return DefaultHandlers.GuildMember(payload, data)
	}
	return nil
}

func messageHandler(payload *dto.Payload, message []byte) error {
	data := &dto.MessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.Message != nil {
		return DefaultHandlers.Message(payload, data)
	}
	return nil
}

func messageDeleteHandler(payload *dto.Payload, message []byte) error {
	data := &dto.MessageDeleteData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.MessageDelete != nil {
		return DefaultHandlers.MessageDelete(payload, data)
	}
	return nil
}

func messageReactionHandler(payload *dto.Payload, message []byte) error {
	data := &dto.MessageReactionData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.MessageReaction != nil {
		return DefaultHandlers.MessageReaction(payload, data)
	}
	return nil
}

func atMessageHandler(payload *dto.Payload, message []byte) error {
	data := &dto.ATMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.ATMessage != nil {
		return DefaultHandlers.ATMessage(payload, data)
	}
	return nil
}

func groupAtMessageHandler(payload *dto.Payload, message []byte) error {
	data := &dto.GroupATMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if _, loaded := processedIDs.LoadOrStore(data.ID, struct{}{}); loaded {
		return nil
	}
	if DefaultHandlers.GroupATMessage != nil {
		return DefaultHandlers.GroupATMessage(payload, data)
	}
	return nil
}

func c2cMessageHandler(payload *dto.Payload, message []byte) error {
	data := &dto.C2CMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.C2CMessage != nil {
		return DefaultHandlers.C2CMessage(payload, data)
	}
	return nil
}

func groupaddbothandler(payload *dto.Payload, message []byte) error {
	data := &dto.GroupAddBotEvent{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.GroupAddbot != nil {
		return DefaultHandlers.GroupAddbot(payload, data)
	}
	return nil
}

func groupdelbothandler(payload *dto.Payload, message []byte) error {
	data := &dto.GroupAddBotEvent{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.GroupDelbot != nil {
		return DefaultHandlers.GroupDelbot(payload, data)
	}
	return nil
}

func publicMessageDeleteHandler(payload *dto.Payload, message []byte) error {
	data := &dto.PublicMessageDeleteData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.PublicMessageDelete != nil {
		return DefaultHandlers.PublicMessageDelete(payload, data)
	}
	return nil
}

func directMessageHandler(payload *dto.Payload, message []byte) error {
	data := &dto.DirectMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.DirectMessage != nil {
		return DefaultHandlers.DirectMessage(payload, data)
	}
	return nil
}

func directMessageDeleteHandler(payload *dto.Payload, message []byte) error {
	data := &dto.DirectMessageDeleteData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.DirectMessageDelete != nil {
		return DefaultHandlers.DirectMessageDelete(payload, data)
	}
	return nil
}

func audioHandler(payload *dto.Payload, message []byte) error {
	data := &dto.AudioData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.Audio != nil {
		return DefaultHandlers.Audio(payload, data)
	}
	return nil
}

func threadHandler(payload *dto.Payload, message []byte) error {
	data := &dto.ThreadData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.Thread != nil {
		return DefaultHandlers.Thread(payload, data)
	}
	return nil
}

func postHandler(payload *dto.Payload, message []byte) error {
	data := &dto.PostData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.Post != nil {
		return DefaultHandlers.Post(payload, data)
	}
	return nil
}

func replyHandler(payload *dto.Payload, message []byte) error {
	data := &dto.ReplyData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.Reply != nil {
		return DefaultHandlers.Reply(payload, data)
	}
	return nil
}

func forumAuditHandler(payload *dto.Payload, message []byte) error {
	data := &dto.ForumAuditData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.ForumAudit != nil {
		return DefaultHandlers.ForumAudit(payload, data)
	}
	return nil
}

func messageAuditHandler(payload *dto.Payload, message []byte) error {
	data := &dto.MessageAuditData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.MessageAudit != nil {
		return DefaultHandlers.MessageAudit(payload, data)
	}
	return nil
}

func interactionHandler(payload *dto.Payload, message []byte) error {
	data := &dto.InteractionEventData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.Interaction != nil {
		return DefaultHandlers.Interaction(payload, data)
	}
	return nil
}

func groupMsgRejecthandler(payload *dto.Payload, message []byte) error {
	data := &dto.GroupMsgRejectEvent{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.GroupMsgReject != nil {
		return DefaultHandlers.GroupMsgReject(payload, data)
	}
	return nil
}

func groupMsgReceivehandler(payload *dto.Payload, message []byte) error {
	data := &dto.GroupMsgReceiveEvent{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.GroupMsgReceive != nil {
		return DefaultHandlers.GroupMsgReceive(payload, data)
	}
	return nil
}
