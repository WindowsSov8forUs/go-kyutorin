package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/echo"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	satoriMessage "github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

// MessageListRequest 获取消息列表请求
type MessageListRequest struct {
	ChannelId string `json:"channel_id"`     // 频道 ID
	Next      string `json:"next,omitempty"` // 分页令牌
}

// MessageListResponse 获取消息列表响应
type MessageListResponse satoriMessage.MessageList

// HandleMessageList 处理获取消息列表请求
func HandleMessageList(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request MessageListRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response MessageListResponse
		var dtoMessages []*dto.Message
		dtoMessages, err = apiv2.Messages(context.TODO(), request.ChannelId, createMessagesPager(request.Next))
		if err != nil {
			return "", err
		}
		response.Next = dtoMessages[len(dtoMessages)-1].ID
		for _, dtoMessage := range dtoMessages {
			m := satoriMessage.Message{
				Id:      dtoMessage.ID,
				Content: processor.ConvertToMessageContent(dtoMessage),
				Channel: &channel.Channel{
					Id: dtoMessage.ChannelID,
				},
				Guild: &guild.Guild{
					Id: dtoMessage.GuildID,
				},
				Member: &guildmember.GuildMember{
					Nick: dtoMessage.Member.Nick,
				},
				User: &user.User{
					Id:     dtoMessage.Author.ID,
					Name:   dtoMessage.Author.Username,
					Avatar: dtoMessage.Author.Avatar,
					IsBot:  dtoMessage.Author.Bot,
				},
			}

			if dtoMessage.DirectMessage {
				m.Channel.Type = channel.CHANNEL_TYPE_DIRECT
			} else {
				m.Channel.Type = channel.CHANNEL_TYPE_TEXT
			}

			time, err := dtoMessage.Member.JoinedAt.Time()
			if err == nil {
				m.Member.JoinedAt = time.Unix()
			}

			time, err = dtoMessage.Timestamp.Time()
			if err == nil {
				m.CreateAt = time.Unix()
			}

			response.Data = append(response.Data, m)
		}

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	} else if message.Platform == "qq" {
		var response MessageListResponse

		// 获取子频道类型
		channelType := echo.GetOpenIdType(request.ChannelId)
		if channelType != "private" {
			channelType = "group"
		}

		messages, next, err := database.GetMessageList(request.ChannelId, channelType, request.Next)
		if err != nil {
			return "", err
		}
		for _, message := range messages {
			response.Data = append(response.Data, *message)
		}
		response.Next = next

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}

// createMessagesPager 构建消息列表范围
func createMessagesPager(next string) *dto.MessagesPager {
	return &dto.MessagesPager{
		Type:  dto.MPTAfter,
		ID:    next,
		Limit: "20",
	}
}
