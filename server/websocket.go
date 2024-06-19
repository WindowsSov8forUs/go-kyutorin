package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/operation"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket WebSocket 服务器
type WebSocket struct {
	conn     *websocket.Conn
	token    string
	mutex    *sync.Mutex
	isClosed chan bool
}

// 定义升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler 对外暴露的 WebSocket 处理函数
func (server *Server) WebSocketHandler(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		webSocketHandler(token, server, c)
	}
}

// webSocketHandler 处理 WebSocket 连接
func webSocketHandler(token string, server *Server, c *gin.Context) {
	// 升级 HTTP 请求为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("建立 WebSocket 服务器时出错: %v", err)
		return
	}
	log.Infof("已建立与 Satori 应用的 WebSocket 连接，IP: %s", c.ClientIP())

	// 创建 WebSocket
	ws := &WebSocket{
		conn:     conn,
		token:    token,
		mutex:    &sync.Mutex{},
		isClosed: make(chan bool),
	}
	// 添加到 server 中
	server.mutex.Lock()
	server.websockets = append(server.websockets, ws)
	server.mutex.Unlock()

	// 在 defer 语句前运行
	defer func() {
		// 从 server 中移除
		server.mutex.Lock()
		for i, v := range server.websockets {
			if v == ws {
				server.websockets = append(server.websockets[:i], server.websockets[i+1:]...)
				break
			}
		}
		server.mutex.Unlock()
		log.Infof("已断开与 Satori 应用的 WebSocket 连接，IP: %s", c.ClientIP())
	}()
	// 关闭连接
	defer ws.conn.Close()

	// 开始鉴权流程
	var sequence int64
	operationChan := make(chan operation.Operation)
	// 开始一个 10s 的计时器
	timer := time.NewTimer(10 * time.Second)
	for {
		// 启动一个一次性接收信令的协程
		go ws.receiveAtOnce(operationChan)
		// 判断接收到的信令类型
		select {
		case sgnl := <-operationChan:
			if sgnl.Op == operation.OpCodeIdentify {
				// 鉴权
				body, err := json.Marshal(sgnl.Body)
				if err != nil {
					continue
				}
				var identify operation.IdentifyBody
				if err := json.Unmarshal(body, &identify); err != nil {
					continue
				}
				log.Infof("收到鉴权信令，鉴权令牌: %s", identify.Token)
				if !ws.authorize(identify.Token) {
					// 鉴权失败
					log.Warn("鉴权失败，请重新进行鉴权")
					continue
				}
				// 鉴权成功
				log.Info("鉴权成功，开始进行事件推送")
				sequence = identify.Sequence
				// 发送 READY 信令
				readyBody := processor.GetReadyBody()
				readyOperation := operation.Operation{
					Op:   operation.OpCodeReady,
					Body: readyBody,
				}
				// 转换为 []byte 并发送
				message, err := json.Marshal(readyOperation)
				if err != nil {
					log.Fatalf("发送 READY 信令时出错: %v", err)
					return
				}
				if err := ws.SendMessage(message); err != nil {
					log.Fatalf("发送 READY 信令时出错: %v", err)
					return
				}
				// 关闭计时器
				timer.Stop()
			}
		case <-timer.C:
			// 10s 计时器到时，终止连接
			log.Warn("鉴权超时，本次连接中断")
			return
		case <-ws.isClosed:
			// 收到关闭信号，终止连接
			return
		}
		break
	}

	// 进行事件补发
	if sequence > 0 {
		// 处理事件队列
		events := server.events.ResumeEvents(sequence)

		if len(events) > 0 {
			log.Infof("开始进行事件补发，起始序列号: %d", sequence)

			// 循环补发事件直到队列清空
			for _, event := range events {
				// 构建 WebSocket 信令
				sgnl := &operation.Operation{
					Op:   operation.OpCodeEvent,
					Body: (*operation.EventBody)(event),
				}
				// 转换为 []byte
				data, err := json.Marshal(sgnl)
				if err != nil {
					log.Errorf("转换信令时出错: %v", err)
					continue
				}
				if err := ws.SendMessage(data); err != nil {
					log.Errorf("补发事件时出错: %v", err)
				}
			}
		}
	}

	// 监听心跳
	go ws.listenHeartbeat()

	<-ws.isClosed
}

// receive 持续接收信令直到接收到关闭信号
func (ws *WebSocket) receive(operationChan chan operation.Operation, errChan chan error) {
	for {
		// 读取信令
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			errChan <- err
			ws.Close()
			return
		}
		// 解析信令
		var op operation.Operation
		if err := json.Unmarshal(message, &op); err != nil {
			continue
		}
		// 发送信令
		operationChan <- op
	}
}

// receiveAtOnce 接收一次信令
func (ws *WebSocket) receiveAtOnce(operationChan chan operation.Operation) {
	for {
		// 读取信令
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			log.Errorf("读取信令时出错: %v", err)
			ws.Close()
			return
		}
		// 解析信令
		var op operation.Operation
		if err := json.Unmarshal(message, &op); err != nil {
			continue
		}
		// 发送信令
		operationChan <- op
		return
	}
}

// listenHeartbeat 监听心跳
func (ws *WebSocket) listenHeartbeat() {
	// 启动信令接收协程
	opChan := make(chan operation.Operation)
	errChan := make(chan error)
	// 开始一个 11s 的计时器
	timer := time.NewTimer(11 * time.Second)
	go ws.receive(opChan, errChan)
	// 判断接收到的信令类型
	for {
		select {
		case sgnl := <-opChan:
			if sgnl.Op == operation.OpCodePing {
				// 收到心跳信令，回复心跳信令
				operationPong := operation.Operation{
					Op: operation.OpCodePong,
				}
				message, err := json.Marshal(operationPong)
				if err != nil {
					log.Errorf("回复心跳信令时出错: %v", err)
				}
				if err := ws.SendMessage(message); err != nil {
					log.Errorf("回复心跳信令时出错: %v", err)
				}
				// 重置计时器
				timer.Reset(11 * time.Second)
			}
		case <-timer.C:
			// 11s 计时器到时，终止连接
			log.Warn("心跳超时，本次连接中断")
			ws.isClosed <- true
			return
		case <-errChan:
			// 读取信令时出错，终止连接
			ws.isClosed <- true
			return
		}
	}
}

// SendMessage 发送消息
func (ws *WebSocket) SendMessage(message []byte) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.conn.WriteMessage(websocket.TextMessage, message)
}

// PostEvent 推送事件
func (ws *WebSocket) PostEvent(event *operation.Event) error {
	op := &operation.Operation{
		Op:   operation.OpCodeEvent,
		Body: (*operation.EventBody)(event),
	}
	message, err := json.Marshal(op)
	if err != nil {
		log.Errorf("转换信令时出错: %v", err)
		return nil
	}
	return ws.SendMessage(message)
}

// Close 关闭 WebSocket 连接
func (ws *WebSocket) Close() {
	// 发送关闭信号
	ws.isClosed <- true
}

// authorize 鉴权
func (ws *WebSocket) authorize(authorization string) bool {
	// 如果设置的令牌为空则默认不进行鉴权
	if ws.token == "" {
		return true
	}

	// 如果设置的令牌不为空则进行鉴权
	return authorization == ws.token
}
