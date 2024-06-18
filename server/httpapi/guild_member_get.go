package httpapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("guild.member.get", HandleGuildMemberGet)
}

// RequestGuildMemberGet 获取群组成员请求
type RequestGuildMemberGet struct {
	GuildId string `json:"guild_id"` // 群组 ID
	UserId  string `json:"user_id"`  // 用户 ID
}

// ResponseGuildMemberGet 获取群组成员响应
type ResponseGuildMemberGet guildmember.GuildMember

// HandleGuildMemberGet 处理获取群组成员请求
func HandleGuildMemberGet(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildMemberGet
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ResponseGuildMemberGet

		var dtoMember *dto.Member
		dtoMember, err = apiv2.GuildMember(context.TODO(), request.GuildId, request.UserId)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		// 将 dto.Member 转换为 guildmember.GuildMember
		guildMember, err := convertDtoMemberToGuildMember(dtoMember)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		response = ResponseGuildMemberGet(guildMember)

		return response, nil
	}

	return defaultResource(message)
}

// convertDtoMemberToGuildMember 将 dto.Member 转换为 guildmember.GuildMember
func convertDtoMemberToGuildMember(dtoMember *dto.Member) (guildmember.GuildMember, error) {
	var guildMember guildmember.GuildMember
	var user = &user.User{}

	// 获取转换后的时间戳
	var joinedAt int64
	joinedAt, err := convertDtoTimestampToInt64(&dtoMember.JoinedAt)
	if err != nil {
		return guildMember, err
	}

	user.Id = dtoMember.User.ID
	user.Name = dtoMember.User.Username
	user.IsBot = dtoMember.User.Bot
	guildMember.Nick = dtoMember.Nick
	guildMember.Avatar = dtoMember.User.Avatar
	guildMember.JoinedAt = joinedAt
	guildMember.User = user

	return guildMember, nil
}

// convertDtoTimestampToInt64 将 dto.Timestamp 转换为 int64
func convertDtoTimestampToInt64(timestamp *dto.Timestamp) (int64, error) {
	if timestamp == nil {
		return 0, fmt.Errorf("empty timestamp.")
	}
	time, err := timestamp.Time()
	if err != nil {
		return 0, err
	}
	return time.Unix(), nil
}
