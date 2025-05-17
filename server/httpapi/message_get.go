package httpapi

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/gin-gonic/gin"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("message.get", HandleMessageGet)
}

// RequestMessageGet 获取消息请求
type RequestMessageGet struct {
	ChannelId string `json:"channel_id"` // 频道 ID
	MessageId string `json:"message_id"` // 消息 ID
}

// ResponseMessageGet 获取消息响应
type ResponseMessageGet message.Message

// HandleMessageGet 处理获取消息请求
func HandleMessageGet(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestMessageGet
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ResponseMessageGet
		var dtoMessage *dto.Message
		dtoMessage, err = apiv2.Message(context.TODO(), request.ChannelId, request.MessageId)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		response.Id = dtoMessage.ID
		response.Content = processor.ConvertToMessageContent(dtoMessage)

		response.Channel = &channel.Channel{
			Id: dtoMessage.ChannelID,
		}
		if dtoMessage.DirectMessage {
			response.Channel.Type = channel.ChannelTypeDirect
		} else {
			response.Channel.Type = channel.ChannelTypeText
		}

		response.Guild = &guild.Guild{
			Id: dtoMessage.GuildID,
		}

		response.Member = &guildmember.GuildMember{
			Nick: dtoMessage.Member.Nick,
		}
		time, err := dtoMessage.Member.JoinedAt.Time()
		if err == nil {
			response.Member.JoinedAt = time.UnixMilli()
		}

		response.User = &user.User{
			Id:     dtoMessage.Author.ID,
			Name:   dtoMessage.Author.Username,
			Avatar: dtoMessage.Author.Avatar,
			IsBot:  dtoMessage.Author.Bot,
		}

		time, err = dtoMessage.Timestamp.Time()
		if err == nil {
			response.CreateAt = time.UnixMilli()
		}

		time, err = dtoMessage.EditedTimestamp.Time()
		if err == nil {
			response.UpdateAt = time.UnixMilli()
		}

		return response, nil
	} else if message.Platform == "qq" {
		var response ResponseMessageGet

		// 获取子频道类型
		channelType := processor.GetOpenIdType(request.ChannelId)
		if channelType != "private" {
			channelType = "group"
		}

		msg, err := database.GetMessage(request.ChannelId, channelType, request.MessageId)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		response = ResponseMessageGet(*msg)

		return response, nil
	}

	return defaultResource(message)
}
