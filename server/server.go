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
			continue
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
	rwMutex    sync.RWMutex
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
	)

	webSocketGroup := engine.Group(fmt.Sprintf("%s/v1/events", server.conf.Satori.Path))
	// WebSocket 处理函数
	webSocketGroup.GET("", server.WebSocketHandler(server.conf.Satori.Token))

	resourceGroup := engine.Group(fmt.Sprintf("%s/v1/", server.conf.Satori.Path))
	// 资源接口处理函数
	resourceGroup.Use(
		httpapi.HeadersValidateMiddleware(),
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
	adminGroup.Use(httpapi.HeadersValidateMiddleware())
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
		rwMutex:    sync.RWMutex{},
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
	log.Debugf("在推送事件时尝试获取读锁")
	server.rwMutex.RLock()
	log.Debugf("在推送事件时获取读锁成功")

	server.events.PushEvent(event)

	var waitGroup sync.WaitGroup

	wsResults := make(chan *WebSocket, len(server.websockets))
	for _, ws := range server.websockets {
		waitGroup.Add(1)
		go func(ws *WebSocket) {
			defer waitGroup.Done()
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
		waitGroup.Add(1)
		go func(wh *WebHook) {
			defer waitGroup.Done()
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
	log.Debugf("等待 %d 个 WebSocket 和 WebHook 客户端推送事件...", len(server.websockets)+len(server.webhooks))
	waitGroup.Wait()
	log.Debug("所有 WebSocket 和 WebHook 客户端已推送事件完成")
	close(wsResults)
	close(whResults)

	server.rwMutex.RUnlock()
	log.Debugf("在推送事件时释放读锁")

	websockets := make([]*WebSocket, 0)
	for ws := range wsResults {
		if ws != nil {
			websockets = append(websockets, ws)
		}
	}
	log.Debugf("WebSocket 存活：(%v/%v)", len(websockets), len(server.websockets))

	webhooks := make([]*WebHook, 0)
	for wh := range whResults {
		if wh != nil {
			webhooks = append(webhooks, wh)
		}
	}

	log.Debug("在统计有效连接时尝试获取写锁")
	server.rwMutex.Lock()
	log.Debug("在统计有效连接时获取写锁成功")
	defer func() {
		log.Debug("在统计有效连接时释放写锁")
		server.rwMutex.Unlock()
	}()

	server.websockets = websockets
	server.webhooks = webhooks
}

func (server *Server) Close() {
	log.Info("正在关闭 Satori 服务端...")

	totalWebSocket := len(server.websockets)
	for index, ws := range server.websockets {
		if ws != nil {
			log.Debugf("正在关闭 WebSocket 连接 (%v/%v) ：%s", index+1, totalWebSocket, ws.IP)
			ws.Close()
			log.Debugf("WebSocket 连接 (%v/%v) 已关闭：%s", index+1, totalWebSocket, ws.IP)
		}
	}

	log.Debug("在关闭服务端时尝试获取写锁")
	server.rwMutex.Lock()
	log.Debug("在关闭服务端时获取写锁成功")
	defer func() {
		log.Debug("在关闭服务端时释放写锁")
		server.rwMutex.Unlock()
	}()

	server.websockets = make([]*WebSocket, 0)
	server.webhooks = make([]*WebHook, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Debug("正在关闭 HTTP 服务器...")
	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.Errorf("关闭 HTTP 服务器时出错: %v", err)
	}

	log.Info("Satori 服务端已关闭")
}
