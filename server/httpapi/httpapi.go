package httpapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/WindowsSov8forUs/go-kyutorin/version"
	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/openapi"
)

// webHookServerManager WebHook 服务端管理器
type webHookServerManager interface {
	CreateWebHook(url, token string) error
	DeleteWebHook(url string) error
}

// Server HTTP 服务端
type Server struct {
	httpServer     *http.Server
	webHookManager webHookServerManager
}

var instance *Server

func (server *Server) Run() error {
	return server.httpServer.ListenAndServe()
}

func (server *Server) Shutdown(ctx context.Context) error {
	return server.httpServer.Shutdown(ctx)
}

func (server *Server) Addr() string {
	return server.httpServer.Addr
}

func NewHttpServer(addr string, handler http.Handler, webHookManager webHookServerManager) *Server {
	instance = &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		webHookManager: webHookManager,
	}
	return instance
}

// APIError API 错误
type APIError interface {
	Error() string
	Code() int
}

// BadRequestError 请求格式错误
type BadRequestError struct {
	err error
}

func (e *BadRequestError) Error() string {
	return e.err.Error()
}

func (e *BadRequestError) Code() int {
	return http.StatusBadRequest
}

// UnauthorizedError 缺失鉴权
type UnauthorizedError struct {
	invalidToken string
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized with wrong token: %s", e.invalidToken)
}

func (e *UnauthorizedError) Code() int {
	return http.StatusUnauthorized
}

// ForbiddenError 权限不足
type ForbiddenError struct {
	message string
}

func (e *ForbiddenError) Error() string {
	return e.message
}

func (e *ForbiddenError) Code() int {
	return http.StatusForbidden
}

// NotFoundError 资源不存在
type NotFoundError struct {
	api      string
	platform string
}

func (e *NotFoundError) Error() string {
	if e.platform == "" {
		return fmt.Sprintf(`api "%s" not found`, e.api)
	} else {
		return fmt.Sprintf(`api "%s" is not supported on %s`, e.api, e.platform)
	}
}

func (e *NotFoundError) Code() int {
	return http.StatusNotFound
}

// MethodNotAllowedError 请求方法不支持
type MethodNotAllowedError struct {
	method string
}

func (e *MethodNotAllowedError) Error() string {
	return fmt.Sprintf(`method "%s" not allowed`, e.method)
}

func (e *MethodNotAllowedError) Code() int {
	return http.StatusMethodNotAllowed
}

// InternalServerError 服务器内部错误
type InternalServerError struct {
	err error
}

func (e *InternalServerError) Error() string {
	return e.err.Error()
}

func (e *InternalServerError) Code() int {
	return http.StatusInternalServerError
}

// ActionMessage Satori 应用发送的 HTTP API 调用信息
type ActionMessage struct {
	API      string       // 接口
	Bot      *user.User   // 机器人信息
	Platform string       // 平台
	Ctx      *gin.Context // 上下文
}

// Data 获取数据
func (message *ActionMessage) Data() []byte {
	body := message.Ctx.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		message.Ctx.String(http.StatusBadRequest, err.Error())
		return nil
	}
	return bodyBytes
}

// MetaActionMessage Satori 应用发送的元信息接口调用信息
type MetaActionMessage struct {
	API string       // 接口
	Ctx *gin.Context // 上下文
}

// Data 获取数据
func (message *MetaActionMessage) Data() []byte {
	body := message.Ctx.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		message.Ctx.String(http.StatusBadRequest, err.Error())
		return nil
	}
	return bodyBytes
}

// NewActionMessage 创建一个新的 ActionMessage
func NewActionMessage(api string, bot *user.User, platform string, ctx *gin.Context) *ActionMessage {
	return &ActionMessage{
		API:      api,
		Bot:      bot,
		Platform: platform,
		Ctx:      ctx,
	}
}

// NewMetaActionMessage 创建一个新的 MetaActionMessage
func NewMetaActionMessage(api string, ctx *gin.Context) *MetaActionMessage {
	return &MetaActionMessage{
		API: "meta" + api,
		Ctx: ctx,
	}
}

// HTTP API 处理函数
type HandlerFunc func(api, apiV2 openapi.OpenAPI, action *ActionMessage) (any, APIError)

// 元信息接口的处理函数
type MetaHandlerFunc func(action *MetaActionMessage) (any, APIError)

var handlers = make(map[string]HandlerFunc)
var metaHandlers = make(map[string]MetaHandlerFunc)

// defaultResource 资源默认处理函数
func defaultResource(action *ActionMessage) (any, APIError) {
	return gin.H{}, &NotFoundError{action.API, action.Platform}
}

// RegisterHandler 注册特定资源与方法的处理函数
func RegisterHandler(api string, handler HandlerFunc) {
	handlers[api] = handler
}

// RegisterMetaHandler 注册元信息接口的处理函数
func RegisterMetaHandler(api string, handler MetaHandlerFunc) {
	if api != "" {
		api = "/" + api
	}
	metaHandlers["meta"+api] = handler
}

// CallAPI 调用 Satori API
func CallAPI(api, apiV2 openapi.OpenAPI, action *ActionMessage) (any, APIError) {
	if _, ok := handlers[action.API]; !ok {
		return gin.H{}, &NotFoundError{api: action.API}
	}
	return handlers[action.API](api, apiV2, action)
}

// CallMetaAPI 调用 Satori 元信息 API
func CallMetaAPI(action *MetaActionMessage) (any, APIError) {
	if _, ok := metaHandlers[action.API]; !ok {
		return gin.H{}, &NotFoundError{api: action.API}
	}
	return metaHandlers[action.API](action)
}

// HeadersSetMiddleware 设置响应头中间件
func HeadersSetMiddleware(satoriVersion string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Header("Date", time.Now().Format(time.RFC1123))
		c.Header("Server", fmt.Sprintf("Go-Kyutorin/%s", version.Version))
		c.Header("X-Satori-Protocol", satoriVersion)
		c.Next()
	}
}

// HeadersValidateMiddleware 请求头验证中间件
func HeadersValidateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取请求头
		contentType := c.GetHeader("Content-Type")
		// 提取请求 API ，用以进行特例判断
		method := c.Param("method")

		// 判断请求头错误
		if method == "upload.create" {
			// 对于文件上传 API ， Content-Type 必须为 multipart/form-data
			if contentType != "multipart/form-data" {
				c.String(http.StatusBadRequest, "content type must be multipart/form-data")
				c.Abort()
				return
			}
		} else {
			// 对于其他 API ， Content-Type 必须为 application/json
			if contentType != "application/json" {
				c.String(http.StatusBadRequest, "content type must be application/json")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// AuthenticateMiddleware 鉴权中间件
func AuthenticateMiddleware(realm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取 Authorization 请求头
		authorization := c.GetHeader("Authorization")

		// 鉴权
		if ok, err := authorize(authorization); !ok {
			c.Header(
				"WWW-Authenticate",
				fmt.Sprintf(
					`Bearer realm="%s", error="%s", error_description="%s", error_url="%s"`,
					realm,
					err.Error(),
					"authorize failed with token: "+authorization,
					`https://satori.js.org/zh-CN/protocol/api.html#%E9%89%B4%E6%9D%83`,
				),
			)
			c.String(http.StatusUnauthorized, "authorize failed with token: "+authorization)
			c.Abort()
			return
		}

		c.Next()
	}
}

// BotValidateMiddleware 机器人验证中间件
func BotValidateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取请求头
		satoriPlatform := c.GetHeader("Satori-Platform")
		satoriUserID := c.GetHeader("Satori-User-ID")

		// 判断平台与 UserID 是否正确
		bot := processor.GetBot(satoriPlatform)
		if bot == nil {
			c.String(http.StatusBadRequest, `unknown platform "%s"`, satoriPlatform)
			c.Abort()
			return
		}
		if satoriUserID != bot.Id {
			c.String(http.StatusBadRequest, `unknown user id "%s"`, satoriUserID)
			c.Abort()
			return
		}

		c.Next()
	}
}

// ResourceMiddleware 资源中间件
func ResourceMiddleware(api, apiV2 openapi.OpenAPI) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 在内部进行判断处理
		resourceAPIHandler(ctx, api, apiV2)
	}
}

// resourceAPIHandler 处理资源 API
func resourceAPIHandler(c *gin.Context, api, apiV2 openapi.OpenAPI) {
	// 提取路径中参数
	method := c.Param("method")

	// 获取 bot 对象
	satoriPlatform := c.GetHeader("Satori-Platform")
	bot := processor.GetBot(satoriPlatform)

	// 构建 Action
	actionMessage := NewActionMessage(method, bot, satoriPlatform, c)

	// 调用 API
	response, err := CallAPI(api, apiV2, actionMessage)
	if err != nil {
		switch err.(type) {
		case *BadRequestError:
			c.String(http.StatusBadRequest, err.Error())
		case *UnauthorizedError:
			c.String(http.StatusUnauthorized, err.Error())
		case *ForbiddenError:
			c.String(http.StatusForbidden, err.Error())
		case *NotFoundError:
			c.String(http.StatusNotFound, err.Error())
		case *MethodNotAllowedError:
			c.String(http.StatusMethodNotAllowed, err.Error())
		case *InternalServerError:
			c.String(http.StatusInternalServerError, err.Error())
		default:
			c.String(http.StatusInternalServerError, err.Error())
		}
	} else {
		// 返回结果
		c.JSON(http.StatusOK, response)
	}
}

// AdminMiddleware 管理接口中间件
func MetaMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 在内部进行判断处理
		metaAPIHandler(ctx)
	}
}

// metaAPIHandler 处理元信息 API
func metaAPIHandler(c *gin.Context) {
	// 提取路径中参数
	method := c.Param("method")

	// 构建 Action
	metaActionMessage := NewMetaActionMessage(method, c)

	// 调用 API
	response, err := CallMetaAPI(metaActionMessage)
	if err != nil {
		switch err.(type) {
		case *BadRequestError:
			c.String(http.StatusBadRequest, err.Error())
		case *UnauthorizedError:
			c.String(http.StatusUnauthorized, err.Error())
		case *ForbiddenError:
			c.String(http.StatusForbidden, err.Error())
		case *NotFoundError:
			c.String(http.StatusNotFound, err.Error())
		case *MethodNotAllowedError:
			c.String(http.StatusMethodNotAllowed, err.Error())
		case *InternalServerError:
			c.String(http.StatusInternalServerError, err.Error())
		default:
			c.String(http.StatusInternalServerError, err.Error())
		}
	} else {
		// 返回结果
		c.JSON(http.StatusOK, response)
	}
}

// authorize 鉴权
func authorize(authorization string) (bool, error) {
	// 获取令牌
	token := config.GetSatoriToken()

	// 如果令牌设置为空则不鉴权
	if token == "" {
		return true, nil
	}

	// 构建并比对鉴权
	if !strings.HasPrefix(authorization, "Bearer ") {
		return false, fmt.Errorf("invalid_request")
	} else if authorization != "Bearer "+token {
		return false, fmt.Errorf("invalid_token")
	} else {
		return true, nil
	}
}
