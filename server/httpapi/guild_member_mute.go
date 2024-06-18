package httpapi

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("guild.member.mute", HandleGuildMemberMute)
}

// RequestGuildMemberMute 禁言群组成员请求
type RequestGuildMemberMute struct {
	GuildId  string `json:"guild_id"` // 群组 ID
	UserId   string `json:"user_id"`  // 用户 ID
	Duration int    `json:"duration"` // 禁言时长 (毫秒)
}

// HandleGuildMemberMute 处理禁言群组成员请求
func HandleGuildMemberMute(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildMemberMute
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		err = api.MemberMute(context.TODO(), request.GuildId, request.UserId, createUpdateGuildMute(&request))
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		return gin.H{}, nil
	}

	return defaultResource(message)
}

// createUpdateGuildMute 创建 dto.UpdateGuildMute
func createUpdateGuildMute(request *RequestGuildMemberMute) *dto.UpdateGuildMute {
	return &dto.UpdateGuildMute{
		MuteSeconds: strconv.Itoa(request.Duration / 1000),
		UserIDs:     []string{request.UserId},
	}
}
