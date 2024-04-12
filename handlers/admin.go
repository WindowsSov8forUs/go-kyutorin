package handlers

import (
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/WindowsSov8forUs/go-kyutorin/webhook"

	"github.com/satori-protocol-go/satori-model-go/pkg/login"
)

// LoginListResponse 获取登录信息列表响应
type LoginListResponse []login.Login

// WebHookCreateRequest 创建 WebHook 请求
type WebHookCreateRequest struct {
	URL   string `json:"url"`             // WebHook 地址
	Token string `json:"token,omitempty"` // 鉴权令牌
}

// WebHookDeleteRequest 移除 WebHook 请求
type WebHookDeleteRequest struct {
	URL string `json:"url"` // WebHook 地址
}

// HandlerLoginList 处理获取登录信息列表请求
func HandlerLoginList(message callapi.AdminMessage) (string, error) {
	var response LoginListResponse

	bots := processor.GetBots()
	for platform, bot := range bots {
		login := login.Login{
			User:     bot,
			SelfId:   processor.SelfId,
			Platform: platform,
			Status:   processor.GetStatus(platform),
		}
		response = append(response, login)
	}

	data, err := json.Marshal(response)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// HandlerWebHookCreate 处理创建 WebHook 请求
func HandlerWebHookCreate(message callapi.AdminMessage) (string, error) {
	var request WebHookCreateRequest
	err := json.Unmarshal([]byte(message.Data), &request)
	if err != nil {
		return "", err
	}

	webhook.CreateWebHook(request.URL, request.Token)

	return "", nil
}

// HandlerWebHookDelete 处理移除 WebHook 请求
func HandlerWebHookDelete(message callapi.AdminMessage) (string, error) {
	var request WebHookDeleteRequest
	err := json.Unmarshal([]byte(message.Data), &request)
	if err != nil {
		return "", err
	}

	webhook.DelWebHook(request.URL)

	return "", nil
}
