package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/gin-gonic/gin"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	satoriMessage "github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("message.list", HandleMessageList)
}

// Direction 消息列表方向
type Direction string

const (
	// DirectionBefore 之前
	DirectionBefore Direction = "before"
	// DirectionAfter 之后
	DirectionAfter Direction = "after"
	// DirectionAround 周围
	DirectionAround Direction = "around"
)

// Order 消息列表排序
type Order string

const (
	// OrderAsc 升序
	OrderAsc Order = "asc"
	// OrderDesc 降序
	OrderDesc Order = "desc"
)

// RequestMessageList 获取消息列表请求
type RequestMessageList struct {
	ChannelId string    `json:"channel_id"`          // 频道 ID
	Next      string    `json:"next,omitempty"`      // 分页令牌
	Direction Direction `json:"direction,omitempty"` // 方向
	Limit     int       `json:"limit,omitempty"`     // 数量
	Order     Order     `json:"order,omitempty"`     // 排序
}

// ResponseMessageList 获取消息列表响应
type ResponseMessageList satoriMessage.MessageBidiList

// HandleMessageList 处理获取消息列表请求
func HandleMessageList(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestMessageList
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if request.Direction == "" {
		request.Direction = DirectionBefore
	}
	if request.Limit == 0 {
		request.Limit = 50
	}
	if request.Order == "" {
		request.Order = OrderAsc
	}

	if request.Next == "" && request.Direction != DirectionBefore {
		return gin.H{}, &BadRequestError{err: errors.New(`"next" is required when "direction" is not "before"`)}
	}

	if message.Platform == "qqguild" {
		var response ResponseMessageList
		var dtoMessages []*dto.Message

		dtoMessages, err = apiv2.Messages(context.TODO(), request.ChannelId, createMessagesPager(&request))
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		// limit := request.Limit
		// if limit > 20 {
		// 	limit = 20
		// }
		// TODO: 对于是否已无更多可获取消息的判断
		switch request.Direction {
		case DirectionBefore:
			response.Next = dtoMessages[0].ID
			response.Prev = dtoMessages[0].ID
		case DirectionAfter:
			response.Next = dtoMessages[len(dtoMessages)-1].ID
			response.Prev = dtoMessages[len(dtoMessages)-1].ID
		case DirectionAround:
			response.Next = dtoMessages[len(dtoMessages)-1].ID
			response.Prev = dtoMessages[0].ID
		}

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
				m.Channel.Type = channel.ChannelTypeDirect
			} else {
				m.Channel.Type = channel.ChannelTypeText
			}

			time, err := dtoMessage.Member.JoinedAt.Time()
			if err == nil {
				m.Member.JoinedAt = time.UnixMilli()
			}

			time, err = dtoMessage.Timestamp.Time()
			if err == nil {
				m.CreateAt = time.UnixMilli()
			}

			if request.Order == OrderAsc {
				response.Data = append(response.Data, &m)
			} else {
				response.Data = append([]*satoriMessage.Message{&m}, response.Data...)
			}
		}

		return response, nil
	} else if message.Platform == "qq" {
		var response ResponseMessageList

		// 获取子频道类型
		channelType := processor.GetOpenIdType(request.ChannelId)
		if channelType != "private" {
			channelType = "group"
		}

		var queryDirection database.QueryDirection
		switch request.Direction {
		case DirectionBefore:
			queryDirection = database.QueryDirectionBefore
		case DirectionAfter:
			queryDirection = database.QueryDirectionAfter
		case DirectionAround:
			queryDirection = database.QueryDirectionAround
		}
		prevs, nexts, err := database.GetMessageList(request.ChannelId, channelType, request.Next, queryDirection, request.Limit)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		var responseNext string = ""
		if len(nexts) > 0 && len(nexts) < request.Limit {
			responseNext = nexts[len(nexts)-1].Id
		}
		var responsePrev string = ""
		if len(prevs) > 0 && len(prevs) < request.Limit {
			responsePrev = prevs[len(prevs)-1].Id
		}

		switch request.Direction {
		case DirectionBefore:
			response.Next = responsePrev
			response.Prev = responsePrev
		case DirectionAfter:
			response.Next = responseNext
			response.Prev = responseNext
		case DirectionAround:
			response.Next = responseNext
			response.Prev = responsePrev
		}

		messages := append(prevs, nexts...)

		sort.Slice(messages, func(i, j int) bool {
			if request.Order == OrderAsc {
				return messages[i].CreateAt < messages[j].CreateAt
			}
			return messages[i].CreateAt > messages[j].CreateAt
		})
		response.Data = messages

		return response, nil
	}

	return defaultResource(message)
}

// createMessagesPager 构建消息列表范围
func createMessagesPager(request *RequestMessageList) *dto.MessagesPager {
	var mpt dto.MessagePagerType
	switch request.Direction {
	case DirectionBefore:
		mpt = dto.MPTBefore
	case DirectionAfter:
		mpt = dto.MPTAfter
	case DirectionAround:
		mpt = dto.MPTAround
	}

	return &dto.MessagesPager{
		Type:  mpt,
		ID:    request.Next,
		Limit: strconv.Itoa(request.Limit),
	}
}
