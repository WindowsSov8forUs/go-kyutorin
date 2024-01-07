package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/echo"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"

	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/dezhishen/satori-model-go/pkg/guildmember"
	"github.com/dezhishen/satori-model-go/pkg/message"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("message", "get", HandleMessageGet)
}

// MessageGetRequest 获取消息请求
type MessageGetRequest struct {
	ChannelId string `json:"channel_id"` // 频道 ID
	MessageId string `json:"message_id"` // 消息 ID
}

// MessageGetResponse 获取消息响应
type MessageGetResponse message.Message

// HandleMessageGet 处理获取消息请求
func HandleMessageGet(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request MessageGetRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response MessageGetResponse
		var dtoMessage *dto.Message
		dtoMessage, err = apiv2.Message(context.TODO(), request.ChannelId, request.MessageId)
		if err != nil {
			return "", err
		}
		response.Id = dtoMessage.ID
		response.Content = processor.ConvertToMessageContent(dtoMessage)

		response.Channel = &channel.Channel{
			Id: dtoMessage.ChannelID,
		}
		if dtoMessage.DirectMessage {
			response.Channel.Type = channel.CHANNEL_TYPE_DIRECT
		} else {
			response.Channel.Type = channel.CHANNEL_TYPE_TEXT
		}

		response.Guild = &guild.Guild{
			Id: dtoMessage.GuildID,
		}

		response.Member = &guildmember.GuildMember{
			Nick: dtoMessage.Member.Nick,
		}
		time, err := dtoMessage.Member.JoinedAt.Time()
		if err == nil {
			response.Member.JoinedAt = time.Unix()
		}

		response.User = &user.User{
			Id:     dtoMessage.Author.ID,
			Name:   dtoMessage.Author.Username,
			Avatar: dtoMessage.Author.Avatar,
			IsBot:  dtoMessage.Author.Bot,
		}

		time, err = dtoMessage.Timestamp.Time()
		if err == nil {
			response.CreateAt = time.Unix()
		}

		time, err = dtoMessage.EditedTimestamp.Time()
		if err == nil {
			response.UpdateAt = time.Unix()
		}

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	} else if message.Platform == "qq" {
		var response MessageGetResponse

		// 获取子频道类型
		channelType := echo.GetOpenIdType(request.ChannelId)
		if channelType != "private" {
			channelType = "group"
		}

		msg, err := database.GetMessage(request.ChannelId, channelType, request.MessageId)
		if err != nil {
			return "", err
		}
		response = MessageGetResponse(*msg)

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}
