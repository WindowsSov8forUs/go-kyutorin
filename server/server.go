package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/openapi"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/operation"
	"github.com/WindowsSov8forUs/go-kyutorin/server/httpapi"
)

// EventQueue 事件队列
type EventQueue struct {
	Events []*operation.Event
	mutex  sync.Mutex
}

// PushEvent 推送事件
func (q *EventQueue) PushEvent(event *operation.Event) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for {
		if len(q.Events) < 1000 {
			break
		}
		q.PopEvent()
	}
	q.Events = append(q.Events, event)
}

// PopEvent 弹出事件
func (q *EventQueue) PopEvent() *operation.Event {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.Events) == 0 {
		return nil
	}
	event := q.Events[0]
	q.Events = q.Events[1:]
	return event
}

// ResumeEvents 恢复事件
func (q *EventQueue) ResumeEvents(Sequence int64) []*operation.Event {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	var events []*operation.Event
	var isFound bool = false
	for _, event := range q.Events {
		if event.Id == Sequence {
			isFound = true
		}
		if isFound {
			events = append(events, event)
		}
	}
	return events
}

// Clear 清空事件队列
func (q *EventQueue) Clear() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Events = make([]*operation.Event, 0)
}

// NewEventQueue 创建事件队列
func NewEventQueue() *EventQueue {
	return &EventQueue{
		Events: make([]*operation.Event, 0),
	}
}

type Server struct {
	mutex      sync.Mutex
	websockets []*WebSocket
	webhooks   []*WebHook
	httpServer *httpapi.Server
	conf       *config.Config
	events     *EventQueue
}

func (server *Server) setupV1Engine(api, apiV2 openapi.OpenAPI) *gin.Engine {
	engine := gin.New()
	engine.Use(
		gin.Recovery(),
		httpapi.HeadersSetMiddleware("1.1"),
		httpapi.HeadersValidateMiddleware(),
	)

	webSocketGroup := engine.Group(fmt.Sprintf("%s/v1/events", server.conf.Satori.Path))
	// WebSocket 处理函数
	webSocketGroup.GET("", server.WebSocketHandler(server.conf.Satori.Token))

	resourceGroup := engine.Group(fmt.Sprintf("%s/v1/", server.conf.Satori.Path))
	// 资源接口处理函数
	resourceGroup.Use(
		httpapi.AuthenticateMiddleware("http_api"),
		httpapi.BotValidateMiddleware(),
	)
	resourceGroup.POST(":method", func(c *gin.Context) {
		method := c.Param("method")
		// 将请求输出
		log.Debugf(
			"收到请求: %s %s，请求头：%v，请求体：%v",
			c.Request.Method,
			method,
			c.Request.Header,
			c.Request.Body,
		)
		httpapi.ResourceMiddleware(api, apiV2)(c)
	})

	adminGroup := engine.Group(fmt.Sprintf("%s/v1/admin/", server.conf.Satori.Path))
	// 管理接口处理函数
	adminGroup.POST(":method", func(c *gin.Context) {
		method := c.Param("method")
		// 将请求输出
		log.Debugf(
			"收到请求: /admin/%s %s，请求头：%v，请求体：%v",
			c.Request.Method,
			method,
			c.Request.Header,
			c.Request.Body,
		)
		httpapi.AdminMiddleware()(c)
	})

	return engine
}

func NewServer(api, apiV2 openapi.OpenAPI, conf *config.Config) (*Server, error) {
	server := &Server{
		mutex:      sync.Mutex{},
		websockets: make([]*WebSocket, 0),
		webhooks:   make([]*WebHook, 0),
		httpServer: nil,
		conf:       conf,
		events:     NewEventQueue(),
	}

	switch conf.Satori.Version {
	case 1:
		server.httpServer = httpapi.NewHttpServer(
			fmt.Sprintf("%s:%d", conf.Satori.Server.Host, conf.Satori.Server.Port),
			server.setupV1Engine(api, apiV2),
			server,
		)
		// server.httpServer = &http.Server{
		// 	Addr:    fmt.Sprintf("%s:%d", conf.Satori.Server.Host, conf.Satori.Server.Port),
		// 	Handler: server.setupV1Engine(api, apiV2),
		// }
	default:
		return nil, fmt.Errorf("unknown Satori protocol version: v%d", conf.Satori.Version)
	}

	return server, nil
}

func (server *Server) Run() error {
	log.Infof("Satori 服务器已启动，地址: %s", server.httpServer.Addr())
	err := server.httpServer.Run()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (server *Server) Send(event *operation.Event) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	server.events.PushEvent(event)

	wsResults := make(chan *WebSocket, len(server.websockets))
	for _, ws := range server.websockets {
		go func(ws *WebSocket) {
			err := ws.PostEvent(event)
			if err != nil {
				log.Errorf("WebSocket 推送事件时出错: %v", err)
				ws.Close()
				wsResults <- nil
			} else {
				wsResults <- ws
			}
		}(ws)
	}

	whResults := make(chan *WebHook, len(server.webhooks))
	for _, wh := range server.webhooks {
		go func(wh *WebHook) {
			if wh == nil {
				whResults <- nil
				return
			}
			err := wh.PostEvent(event)
			if err != nil {
				url := wh.GetURL()
				switch err {
				case ErrUnauthorized:
					log.Errorf("WebHook 客户端 %s 鉴权失败，已停止对该 WebHook 客户端的事件推送。", url)
					wh = nil
				case ErrServerError:
					log.Errorf("WebHook 客户端出现内部错误，请检查 WebHook 客户端是否正常。")
				default:
					log.Errorf("向 WebHook 客户端 %s 发送事件时出错: %v", url, err)
					wh = nil
				}
			}
			whResults <- wh
		}(wh)
	}

	// 等待 goroutine 完成
	websockets := make([]*WebSocket, 0)
	for range server.websockets {
		ws := <-wsResults
		if ws != nil {
			websockets = append(websockets, ws)
		}
	}
	server.websockets = websockets

	webhooks := make([]*WebHook, 0)
	for range server.webhooks {
		wh := <-whResults
		if wh != nil {
			webhooks = append(webhooks, wh)
		}
	}
	server.webhooks = webhooks
}

func (server *Server) Close() {
	log.Info(("正在关闭 Satori 服务端..."))

	for _, ws := range server.websockets {
		if ws != nil {
			ws.Close()
		}
	}
	server.websockets = make([]*WebSocket, 0)

	server.mutex.Lock()
	defer server.mutex.Unlock()

	server.webhooks = make([]*WebHook, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.Errorf("关闭 HTTP 服务器时出错: %v", err)
	}
}
