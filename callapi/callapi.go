package callapi

import (
	"errors"

	"github.com/WindowsSov8forUs/go-kyutorin/signaling"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/openapi"
)

// Satori 应用发送的调用信息
type ActionMessage struct {
	resource string    // 资源
	method   string    // 方法
	Bot      user.User // 机器人信息
	Platform string    // 平台
	Data     []byte    // 应用发送的数据
}

// Satori 应用发送的管理接口调用信息
type AdminMessage struct {
	Resource string // 资源
	Method   string // 方法
	Data     []byte // 应用发送的数据
}

// WebSocketServer WebSocket 服务器接口
type WebSocketServer interface {
	SendMessage(message []byte) error
	Close() error
}

// WebHookClient WebHook 客户端接口
type WebHookClient interface {
	PostEvent(*signaling.Event) error
	GetURL() string
}

// 特定资源和方法的处理函数
type HandlerFunc func(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message ActionMessage) (string, error)

// 管理接口的处理函数
type AdminHandlerFunc func(message AdminMessage) (string, error)

var (
	handlers      = make(map[string]map[string]HandlerFunc)
	handlersAdmin = make(map[string]AdminHandlerFunc)
)

var ErrBadRequest = errors.New("bad request")
var ErrUnauthorized = errors.New("unauthorized")
var ErrNotFound = errors.New("not found")
var ErrMethodNotAllowed = errors.New("method not allowed")
var ErrServerError = errors.New("server error")

// RegisterHandler 注册特定资源与方法的处理函数
func RegisterHandler(resource, method string, handler HandlerFunc) {
	if _, ok := handlers[resource]; !ok {
		handlers[resource] = make(map[string]HandlerFunc)
	}
	handlers[resource][method] = handler
}

// CallAPI 调用 Satori API
func CallAPI(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message ActionMessage) (string, error) {
	if _, ok := handlers[message.resource]; !ok {
		return "", ErrNotFound
	}
	if _, ok := handlers[message.resource][message.method]; !ok {
		return "", ErrMethodNotAllowed
	}
	return handlers[message.resource][message.method](api, apiv2, message)
}

// NewActionMessage 创建 ActionMessage
func NewActionMessage(resource string, method string, bot user.User, platform string, data []byte) ActionMessage {
	return ActionMessage{
		resource: resource,
		method:   method,
		Bot:      bot,
		Platform: platform,
		Data:     data,
	}
}

// NewAdminMessage 创建 AdminMessage
func NewAdminMessage(resource, method string, data []byte) AdminMessage {
	return AdminMessage{
		Resource: resource,
		Method:   method,
		Data:     data,
	}
}
