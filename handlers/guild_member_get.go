package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/dezhishen/satori-model-go/pkg/guildmember"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("guild.member", "get", HandleGuildMemberGet)
}

// GuildMemberGetRequest 获取群组成员请求
type GuildMemberGetRequest struct {
	GuildId string `json:"guild_id"` // 群组 ID
	UserId  string `json:"user_id"`  // 用户 ID
}

// GuildMemberGetResponse 获取群组成员响应
type GuildMemberGetResponse guildmember.GuildMember

// HandleGuildMemberGet 处理获取群组成员请求
func HandleGuildMemberGet(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildMemberGetRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response GuildMemberGetResponse

		var dtoMember *dto.Member
		dtoMember, err = apiv2.GuildMember(context.TODO(), request.GuildId, request.UserId)
		if err != nil {
			return "", err
		}

		// 将 dto.Member 转换为 guildmember.GuildMember
		guildMember, err := convertDtoMemberToGuildMember(dtoMember)
		if err != nil {
			return "", err
		}
		response = GuildMemberGetResponse(guildMember)

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}

// convertDtoMemberToGuildMember 将 dto.Member 转换为 guildmember.GuildMember
func convertDtoMemberToGuildMember(dtoMember *dto.Member) (guildmember.GuildMember, error) {
	var guildMember guildmember.GuildMember
	var user *user.User

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
		return 0, fmt.Errorf("空的时间戳")
	}
	time, err := timestamp.Time()
	if err != nil {
		return 0, err
	}
	return time.Unix(), nil
}
