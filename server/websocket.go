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
	IP        string
	conn      *websocket.Conn
	token     string
	mutex     *sync.Mutex
	isClosed  chan bool
	hasClosed chan bool
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
		IP:        c.ClientIP(),
		conn:      conn,
		token:     token,
		mutex:     &sync.Mutex{},
		isClosed:  make(chan bool),
		hasClosed: make(chan bool, 1),
	}

	defer func() {
		ws.hasClosed <- true
	}()

	// 开始鉴权流程
	var sn int64
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
				sn = identify.Sn
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

	// 启动监听心跳
	go ws.listenHeartbeat()

	// 添加到 server 中
	server.rwMutex.Lock()
	server.websockets = append(server.websockets, ws)
	server.rwMutex.Unlock()

	defer func() {
		// 从 server 中移除
		server.rwMutex.Lock()
		for i, v := range server.websockets {
			if v == ws {
				server.websockets = append(server.websockets[:i], server.websockets[i+1:]...)
				break
			}
		}
		server.rwMutex.Unlock()

		// 显式发送关闭帧
		if err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
			log.Debugf("向 %s 发送关闭帧时出错: %v", c.ClientIP(), err)
		}

		// 关闭连接
		err := ws.conn.Close()
		if err != nil {
			log.Debugf("关闭与 %s 的 WebSocket 连接时出错: %v", c.ClientIP(), err)
		}
		log.Infof("已断开与 Satori 应用的 WebSocket 连接，IP: %s", c.ClientIP())
	}()

	// 进行事件补发
	if sn > 0 {
		// 处理事件队列
		events := server.events.ResumeEvents(sn)

		if len(events) > 0 {
			log.Infof("开始进行事件补发，起始序列号: %d", sn)

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
		log.Tracef("收到来自 WebSocket 客户端 (%s) 的信令: %s", ws.IP, message)
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
		return
	}
	// 发送信令
	operationChan <- op
}

// listenHeartbeat 监听心跳
func (ws *WebSocket) listenHeartbeat() {
	// 启动信令接收协程
	opChan := make(chan operation.Operation, 1)
	errChan := make(chan error, 1)
	go ws.receive(opChan, errChan)
	// 开始一个 11s 的计时器
	timer := time.NewTimer(110 * time.Second)
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
	log.Tracef("正在向 WebSocket 客户端 (%s) 发送信令: %s", ws.IP, message)
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
	<-ws.hasClosed
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
