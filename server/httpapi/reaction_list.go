package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("reaction.list", HandleReactionList)
}

// RequestReactionList 获取表态列表请求
type RequestReactionList struct {
	ChannelId string `json:"channel_id"`     // 频道 ID
	MessageId string `json:"message_id"`     // 消息 ID
	Emoji     string `json:"emoji"`          // 表态名称
	Next      string `json:"next,omitempty"` // 分页令牌
}

// ReactionListResponse 获取表态列表响应
type ReactionListResponse user.UserList

// HandleReactionList 处理获取表态列表请求
func HandleReactionList(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestReactionList
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ReactionListResponse
		var dtoMessageReactionUsers *dto.MessageReactionUsers
		dtoEmoji := dto.Emoji{
			ID:   request.Emoji,
			Type: 1, // 统一为 QQ 系统表情
		}
		dtoMessageReactionUsers, err = apiv2.GetMessageReactionUsers(context.TODO(), request.ChannelId, request.MessageId, dtoEmoji, createMessageReactionPager(request.Next))
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		if dtoMessageReactionUsers.IsEnd {
			response.Next = ""
		} else {
			response.Next = dtoMessageReactionUsers.Cookie
		}
		for _, dtoUser := range dtoMessageReactionUsers.Users {
			user := &user.User{
				Id:     dtoUser.ID,
				Name:   dtoUser.Username,
				Avatar: dtoUser.Avatar,
				IsBot:  dtoUser.Bot,
			}
			response.Data = append(response.Data, user)
		}

		return response, nil
	}

	return defaultResource(message)
}

// createMessageReactionPager 构建消息表态列表范围
func createMessageReactionPager(next string) *dto.MessageReactionPager {
	return &dto.MessageReactionPager{
		Cookie: next,
		Limit:  "20",
	}
}
