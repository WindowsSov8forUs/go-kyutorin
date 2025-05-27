package server

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin" // Gin 是一个高性能的 HTTP Web 框架

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/log"
	"github.com/tencent-connect/botgo/webhook"
)

// DefaultQueueSize 监听队列的缓冲长度
const DefaultQueueSize = 10000

// 定义全局变量
var global_s int64

// PayloadWithTimestamp 存储带时间戳的 Payload
type PayloadWithTimestamp struct {
	Payload   *dto.Payload
	Timestamp time.Time
}

var dataMap sync.Map

func init() {
	StartCleanupRoutine()
}

// Setup 依赖注册
func Setup() {
	webhook.Register(&Server{})
}

// WebhookHandler 是处理 webhook 事件的接口
type WebhookHandler interface {
	Handle(c *gin.Context, payload []byte) error
}

func (s *Server) New(config dto.Config) webhook.WebHook {
	engine := gin.New()
	gin.SetMode(gin.DebugMode)
	engine.Use(gin.Recovery())

	return &Server{
		engine: engine,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler: engine,
		},
		messageQueue: make(messageChan, DefaultQueueSize),
		appId:        config.AppId,
		botSecret:    config.BotSecret,
		config:       &config,
	}
}

// Server 是 webhook 服务器的实现
type Server struct {
	engine       *gin.Engine
	server       *http.Server
	messageQueue messageChan
	appId        uint64
	botSecret    string
	config       *dto.Config
}

type messageChan chan *dto.Payload

// Listen 启动 webhook 服务器
func (s *Server) Listen() error {
	// 注册 webhook 路由
	webhookGroup := s.engine.Group(s.config.Path)
	webhookGroup.Use(s.signatureValidateMiddleware())
	webhookGroup.POST("", s.webhookHandler())
	go s.listenMessageAndHandle()

	// 启动 HTTP 服务器
	log.Warnf("由于 Go 已不再支持 SSLv3 证书文件，请务必通过其他方式进行反代，否则无法配置给 QQ 开放平台。")
	log.Infof("启动 HTTP 服务器，地址: %s:%d", s.config.Host, s.config.Port)
	return s.server.ListenAndServe()
}

// Close 停止 webhook 服务器
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// webhookHandler 返回处理 webhook 请求的 gin 处理函数
func (s *Server) webhookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Errorf("读取请求体失败: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read request body"})
			return
		}

		// 预处理请求
		payload, err := s.parseMessageToPayload(body)
		if err != nil {
			log.Errorf("解析请求体失败: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload data"})
			return
		}

		if payload.OPCode == dto.HTTPCallbackValidation {
			// 处理验证请求
			rspBytes, err := s.handleValidation(c, payload)
			if err != nil {
				log.Errorf("处理验证请求失败: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid validation request"})
				return
			}
			c.Data(http.StatusOK, "application/json", rspBytes)
			return
		}

		go s.readMessageToQueue(payload)

		// 总是返回成功
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func (s *Server) signatureValidateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 根据botSecret进行repeat操作后得到seed值计算出公钥
		seed := s.botSecret
		for len(seed) < ed25519.SeedSize {
			seed = strings.Repeat(seed, 2)
		}
		rand := strings.NewReader(seed[:ed25519.SeedSize])
		publicKey, _, err := ed25519.GenerateKey(rand)
		if err != nil {
			log.Errorf("%s ed25519 generate key failed:", s.config, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
			return
		}

		// 取HTTP header中X-Signature-Ed25519(进行hex解码)并校验
		signature := c.GetHeader("X-Signature-Ed25519")
		if signature == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "lack of signature"})
			return
		}
		sig, err := hex.DecodeString(signature)
		if err != nil {
			log.Errorf("%s hex decode signature failed:", s.config, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
		if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
			log.Errorf("%s signature length is not valid:", s.config, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}

		// 取HTTP header中 X-Signature-Timestamp 并校验
		timestamp := c.GetHeader("X-Signature-Timestamp")
		if timestamp == "" {
			log.Errorf("%s X-Signature-Timestamp is empty:", s.config)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "lack of timestamp"})
			return
		}
		httpBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Errorf("%s read http body failed:", s.config, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read body"})
			return
		}
		// 重新设置请求体，以便后续处理
		c.Request.Body = io.NopCloser(bytes.NewBuffer(httpBody))
		// 按照timestamp+Body顺序组成签名体
		var msg bytes.Buffer
		msg.WriteString(timestamp)
		msg.Write(httpBody)

		if ed25519.Verify(publicKey, msg.Bytes(), sig) {
			c.Next()
		} else {
			log.Errorf("%s ed25519 verify failed:", s.config)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}
}

func (s *Server) parseMessageToPayload(message []byte) (*dto.Payload, error) {
	payload := &dto.Payload{}
	if err := json.Unmarshal(message, payload); err != nil {
		log.Errorf("%s json failed, %v", s.config, err)
		return nil, err
	}
	atomic.StoreInt64(&global_s, payload.S)

	payload.RawMessage = message
	log.Infof("%s receive %s message, %s", s.config, dto.OPMeans(payload.OPCode), string(message))
	return payload, nil
}

func (s *Server) readMessageToQueue(payload *dto.Payload) {
	// 计算数据的哈希值
	dataHash := calculateDataHash(payload.Data)

	// 检查是否已存在相同的 Data
	if existingPayload, ok := getDataFromSyncMap(dataHash); ok {
		// 如果已存在相同的 Data，则丢弃当前消息
		log.Infof("%s discard duplicate message with DataHash: %v", s.config, existingPayload)
		return
	}

	// 将新的 payload 存入 sync.Map
	storeDataToSyncMap(dataHash, payload)

	s.messageQueue <- payload
}

func getDataFromSyncMap(dataHash string) (*dto.Payload, bool) {
	value, ok := dataMap.Load(dataHash)
	if !ok {
		return nil, false
	}
	payloadWithTimestamp, ok := value.(*PayloadWithTimestamp)
	if !ok {
		return nil, false
	}
	return payloadWithTimestamp.Payload, true
}

func storeDataToSyncMap(dataHash string, payload *dto.Payload) {
	payloadWithTimestamp := &PayloadWithTimestamp{
		Payload:   payload,
		Timestamp: time.Now(),
	}
	dataMap.Store(dataHash, payloadWithTimestamp)
}

func calculateDataHash(data interface{}) string {
	dataBytes, _ := json.Marshal(data)
	return string(dataBytes) // 这里直接转换为字符串，可以使用更复杂的算法
}

// 在全局范围通过atomic访问s值与message_id的映射
func GetGlobalS() int64 {
	return atomic.LoadInt64(&global_s)
}

func (s *Server) listenMessageAndHandle() {
	defer func() {
		// panic，一般是由于业务自己实现的 handle 不完善导致
		if err := recover(); err != nil {
			log.Errorf("%s listen message and handle panic: %v", s.config, err)
		}
	}()
	for payload := range s.messageQueue {
		go event.ParseAndHandle(payload)
	}
	log.Infof("%s message queue is closed", s.config)
}

func (s *Server) handleValidation(c *gin.Context, payload *dto.Payload) ([]byte, error) {
	appid := c.GetHeader("X-Bot-Appid")
	appidInt, err := strconv.Atoi(appid)
	if err != nil || uint64(appidInt) != s.appId {
		log.Errorf("%s callback address verify appid not match, %s", s.config, appid)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "appid不匹配"})
		return nil, fmt.Errorf("appid不匹配")
	}

	userAgent := c.GetHeader("User-Agent")
	if userAgent != "QQBot-Callback" {
		log.Errorf("%s callback address verify userAgent not match, %s", s.config, userAgent)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userAgent不匹配"})
		return nil, fmt.Errorf("userAgent不匹配")
	}

	validationPayload := &dto.WHValidationRequest{}
	if err := event.ParseData(payload.RawMessage, validationPayload); err != nil {
		log.Errorf("%s callback address verify data parse failed, %v, message %v", s.config, err, payload.RawMessage)
		return nil, fmt.Errorf("data parse failed")
	}
	signature, err := s.calculateSignature(validationPayload)
	if err != nil {
		log.Errorf("%s calculateSignature failed, %v", s.config, err)
		return nil, fmt.Errorf("calculateSignature failed")
	}
	rspBytes, err := json.Marshal(
		&dto.WHValidationResponse{
			PlainToken: validationPayload.PlainToken,
			Signature:  signature,
		},
	)
	if err != nil {
		log.Errorf("handle validation failed:", err)
		return nil, fmt.Errorf("handle validation failed")
	}
	return rspBytes, nil
}

// 计算回调地址验证需要的签名
func (s *Server) calculateSignature(payload *dto.WHValidationRequest) (string, error) {
	seed := s.botSecret
	for len(seed) < ed25519.SeedSize {
		seed = strings.Repeat(seed, 2)
	}
	seed = seed[:ed25519.SeedSize]
	reader := strings.NewReader(seed)
	// GenerateKey 方法会返回公钥、私钥，这里只需要私钥进行签名生成不需要返回公钥
	_, privateKey, err := ed25519.GenerateKey(reader)
	if err != nil {
		log.Errorf("ed25519 generate key failed:", err)
		return "", err
	}
	var msg bytes.Buffer
	msg.WriteString(payload.EventTs)
	msg.WriteString(payload.PlainToken)
	signature := hex.EncodeToString(ed25519.Sign(privateKey, msg.Bytes()))
	return signature, nil
}

const cleanupInterval = 5 * time.Minute // 清理间隔时间

func StartCleanupRoutine() {
	go func() {
		for {
			<-time.After(cleanupInterval)
			cleanupDataMap()
		}
	}()
}

func cleanupDataMap() {
	now := time.Now()
	dataMap.Range(func(key, value interface{}) bool {
		payloadWithTimestamp, ok := value.(*PayloadWithTimestamp)
		if !ok {
			return true
		}

		// 检查时间戳，清理超过一定时间的数据
		if now.Sub(payloadWithTimestamp.Timestamp) > cleanupInterval {
			dataMap.Delete(key)
		}

		return true
	})
}
