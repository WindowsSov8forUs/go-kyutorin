package httpapi

import (
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/gin-gonic/gin"

	"github.com/satori-protocol-go/satori-model-go/pkg/login"
)

func init() {
	RegisterAdminHandler("login.list", HandlerLoginList)
	RegisterAdminHandler("webhook.create", HandlerWebHookCreate)
	RegisterAdminHandler("webhook.delete", HandlerWebHookDelete)
}

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
func HandlerLoginList(message *AdminActionMessage) (any, APIError) {
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

	return response, nil
}

// HandlerWebHookCreate 处理创建 WebHook 请求
func HandlerWebHookCreate(message *AdminActionMessage) (any, APIError) {
	var request WebHookCreateRequest
	err := json.Unmarshal([]byte(message.Data), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	err = instance.webHookManager.CreateWebHook(request.URL, request.Token)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	return gin.H{}, nil
}

// HandlerWebHookDelete 处理移除 WebHook 请求
func HandlerWebHookDelete(message *AdminActionMessage) (any, APIError) {
	var request WebHookDeleteRequest
	err := json.Unmarshal([]byte(message.Data), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	err = instance.webHookManager.DeleteWebHook(request.URL)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	return gin.H{}, nil
}
