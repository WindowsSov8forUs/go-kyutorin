package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/openapi"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/operation"
	"github.com/WindowsSov8forUs/go-kyutorin/server/httpapi"
)

var ErrBadRequest = errors.New("bad request")
var ErrUnauthorized = errors.New("unauthorized")
var ErrNotFound = errors.New("not found")
var ErrMethodNotAllowed = errors.New("method not allowed")
var ErrServerError = errors.New("server error")

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
	rwMutex    sync.RWMutex
	websockets []*WebSocket
	webhooks   []*WebHook
	httpServer *http.Server
	conf       *config.Config
	events     *EventQueue
}

func (server *Server) setupV1Engine(api, apiV2 openapi.OpenAPI) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	group := engine.Group(fmt.Sprintf("%s/v1", server.conf.Satori.Path))
	// WebSocket 处理函数
	group.GET("/events", server.WebSocketHandler(server.conf.Satori.Token))

	// 接口处理函数
	group.POST("/*action", func(c *gin.Context) {
		action := c.Param("action")
		// 将请求输出
		log.Debugf(
			"收到请求: %s %s，请求头：%v，请求体：%v",
			c.Request.Method,
			action,
			c.Request.Header,
			c.Request.Body,
		)
		if strings.HasPrefix(action, "/admin") {
			// 管理接口
			httpapi.AdminMiddleware()(c)
		} else {
			// 资源接口
			httpapi.ResourceMiddleware(api, apiV2)(c)
		}
	})
	return engine
}

func NewServer(api, apiV2 openapi.OpenAPI, conf *config.Config) (*Server, error) {
	switch conf.Satori.Version {
	case 1:
		server := &Server{
			rwMutex:    sync.RWMutex{},
			websockets: make([]*WebSocket, 0),
			webhooks:   make([]*WebHook, 0),
			httpServer: nil,
			conf:       conf,
			events:     NewEventQueue(),
		}
		server.httpServer = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", conf.Satori.Server.Host, conf.Satori.Server.Port),
			Handler: server.setupV1Engine(api, apiV2),
		}
		return server, nil
	default:
		return nil, fmt.Errorf("未知的 Satori 协议版本: v%d", conf.Satori.Version)
	}
}

func (server *Server) Run() error {
	err := server.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (server *Server) Send(event *operation.Event) {
	server.rwMutex.RLock()
	defer server.rwMutex.RUnlock()

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
					log.Errorf("WebHook 服务器 %s 鉴权失败，已停止对该 WebHook 服务器的事件推送。", url)
					wh = nil
				case ErrServerError:
					log.Errorf("WebHook 服务器出现内部错误，请检查 WebHook 服务器是否正常。")
				default:
					log.Errorf("向 WebHook 服务器 %s 发送事件时出错: %v", url, err)
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
	server.rwMutex.Lock()
	defer server.rwMutex.Unlock()

	log.Info(("正在关闭 Satori 服务器..."))

	for _, ws := range server.websockets {
		err := ws.Close()
		log.Warnf("关闭 WebSocket 服务器时出错: %v", err)
	}
	server.websockets = make([]*WebSocket, 0)

	server.webhooks = make([]*WebHook, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.Errorf("关闭 HTTP 服务器时出错: %v", err)
	}
}
