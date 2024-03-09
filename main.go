package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/fileserver"
	"github.com/WindowsSov8forUs/go-kyutorin/httpapi"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"
	"github.com/WindowsSov8forUs/go-kyutorin/sys"
	"github.com/WindowsSov8forUs/go-kyutorin/webhook"
	wsServer "github.com/WindowsSov8forUs/go-kyutorin/websocket"

	"github.com/dezhishen/satori-model-go/pkg/login"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
	"github.com/tencent-connect/botgo/websocket"
)

// 消息处理器，持有 openapi 对象
var p *processor.Processor

func main() {
	// 定义 faststart 命令行标志，默认为 false
	fastStart := flag.Bool("faststart", false, "是否快速启动")

	// 解析命令行参数到定义的标志
	flag.Parse()

	// 检查是否使用了 -faststart 参数
	if !*fastStart {
		sys.InitBase()
	}

	// 检查 config.yml 是否存在
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		var err error
		configData := config.ConfigTemplate

		// 写入 config.yml
		err = os.WriteFile("config.yml", []byte(configData), 0644)
		if err != nil {
			log.Fatalf("写入配置文件时出错: %v", err)
			return
		}

		log.Info("已生成默认配置文件 config.yml，请修改后重启程序")
		fmt.Println("按下任意键继续...")
		fmt.Scanln()
		os.Exit(0)
	}

	// 加载配置
	conf, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("加载配置文件时出错: %v", err)
		return
	}

	// 配置日志等级
	log.SetLogLevel(conf.LogLevel)

	var api openapi.OpenAPI
	var apiV2 openapi.OpenAPI
	var notLogin bool

	// 设置 logger
	logger := log.GetLogger()
	botgo.SetLogger(logger)

	// 如果配置并未设置
	if conf.Account.Token == "" {
		log.Fatal("检测到未完成机器人配置，请修改配置文件后重启程序")
		return
	}

	token := token.BotToken(conf.Account.AppID, conf.Account.AppSecret, conf.Account.Token, token.TypeQQBot)
	ctx := context.Background()
	if err := token.InitToken(ctx); err != nil {
		log.Fatalf("初始化 Token 时出错: %v", err)
		return
	}

	// 如果 intents 为空则抛出错误
	if len(conf.Account.WebSocket.Intents) == 0 {
		log.Fatalf("未设置 intents ，请完成设置后重启程序")
		return
	}

	// 创建 api
	if !conf.Account.Sandbox {
		// 创建 v1 版本 OpenAPI
		if err := botgo.SelectOpenAPIVersion(openapi.APIv1); err != nil {
			log.Fatalf("创建 OpenAPI 时出错: %v", err)
		}
		api = botgo.NewOpenAPI(token).WithTimeout(10 * time.Second)

		// 创建 v2 版本 OpenAPI
		if err := botgo.SelectOpenAPIVersion(openapi.APIv2); err != nil {
			log.Fatalf("创建 OpenAPI 时出错: %v", err)
		}
		apiV2 = botgo.NewOpenAPI(token).WithTimeout(10 * time.Second)
	} else {
		// 创建 v1 版本 OpenAPI
		if err := botgo.SelectOpenAPIVersion(openapi.APIv1); err != nil {
			log.Fatalf("创建 OpenAPI 时出错: %v", err)
		}
		api = botgo.NewSandboxOpenAPI(token).WithTimeout(10 * time.Second)

		// 创建 v2 版本 OpenAPI
		if err := botgo.SelectOpenAPIVersion(openapi.APIv2); err != nil {
			log.Fatalf("创建 OpenAPI 时出错: %v", err)
		}
		apiV2 = botgo.NewSandboxOpenAPI(token).WithTimeout(10 * time.Second)
	}

	var me *dto.User
	me, err = api.Me(ctx)
	if err != nil {
		log.Errorf("获取机器人信息时出错: %v", err)
		notLogin = true
	}
	if !notLogin {
		bot := &user.User{
			Id:     me.ID,
			Name:   me.Username,
			Avatar: me.Avatar,
			IsBot:  me.Bot,
		}
		processor.SetBot("qq", bot)
		processor.SetBot("qqguild", bot)
		processor.SelfId = me.ID

		// 获取 WebSocket 信息
		wsInfo, err := apiV2.WS(ctx, nil, "")
		if err != nil {
			log.Fatalf("获取 WebSocket 信息时出错: %v", err)
		}

		// 定义和初始化 intent
		var intent dto.Intent = 0

		// 动态订阅 intent
		for _, intentName := range conf.Account.WebSocket.Intents {
			handlers, ok := getHandlersByName(intentName)
			if !ok {
				log.Warnf("未知的 intent : %s", intentName)
				continue
			}

			//多次位与 并且订阅事件
			for _, handler := range handlers {
				intent |= websocket.RegisterHandlers(handler)
			}
		}

		log.Infof("注册的 intents : %d", intent)

		// 启动 session manager 管理 websocket 连接
		// Gensokyo 强行设置分片数为 1 了，所以我也这么做吧
		go func() {
			wsInfo.Shards = conf.Account.WebSocket.Shards
			if err = botgo.NewSessionManager().Start(wsInfo, token, &intent); err != nil {
				log.Fatalf("启动 Session Manager 时出错: %v", err)
			}
		}()

		log.Info("已成功连接 QQ 开放平台")
		log.Infof("欢迎使用机器人: %s ！", me.Username)

		// 开启本地文件服务器
		var hasFileServer bool
		if conf.FileServer.UseLocalFileServer {
			if conf.FileServer.URL != "" && conf.FileServer.Port != 0 {
				hasFileServer = true
				fileserver.StartFileServer(conf)
			} else {
				log.Warn("文件服务器 URL 或端口未指定，将不会启动文件服务器")
			}
		}
		if !hasFileServer {
			log.Warn("文件服务器未启动，将无法使用本地文件或 base64 编码发送文件")
		}

		// 启动文件数据库
		if conf.Database.FileDatabase {
			database.StartFileDB()
		} else {
			log.Warn("数据库未启动，将无法使用文件缓存。")
		}

		// 启动消息数据库
		if conf.Database.MessageDatabase.InUse {
			log.Info("正在启动消息数据库...")
			err := database.StartMessageDB(conf.Database.MessageDatabase.Limit)
			if err != nil {
				log.Errorf("启动消息数据库时出错，将无法使用消息缓存: %v", err)
			}
		} else {
			log.Warn("消息数据库未启动，将无法使用消息缓存。")
		}

		p = processor.NewProcessor(api, apiV2)

		// 根据版本设置启动 Satori 服务
		version := fmt.Sprintf("v%d", conf.Satori.Version)

		// 判断 Satori 部署路径是否符合要求
		if conf.Satori.Path != "" && conf.Satori.Path[0] != '/' {
			log.Warnf("Satori 部署路径 %s 不符合要求，将不会作为部署路径添加", conf.Satori.Path)
			conf.Satori.Path = ""
		}

		// 设置 WebHook 超时时间
		webhook.Timeout = conf.Satori.WebHook.Timeout

		r := gin.New()
		r.Use(gin.Recovery())

		// 判断不同版本
		switch version {
		case "v1":
			satoriGroup := r.Group(fmt.Sprintf("%s/%s", conf.Satori.Path, version))
			{
				// 注册 Satori WebSocket 处理函数
				satoriGroup.GET("/events", wsServer.WebSocketHandler(conf.Satori.Token, p))
				// 注册 Satori HTTP API 处理函数
				satoriGroup.POST("/*action", func(c *gin.Context) {
					action := c.Param("action")
					if strings.HasPrefix(action, "/admin") {
						// Satori 管理接口处理函数
						httpapi.AdminMiddleware()(c)
					} else {
						// Satori 资源 API 处理函数
						httpapi.ResourceMiddleware(api, apiV2)(c)
					}
				})
			}
		default:
			log.Fatalf("未知的 Satori 版本: %s", version)
		}

		// 创建一个 http.Server
		httpServer := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", conf.Satori.Server.Host, conf.Satori.Server.Port),
			Handler: r,
		}
		log.Infof("Satori 服务器已启动，地址: %s", httpServer.Addr)

		// 在一个新的 goroutine 中启动 http.Server
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				log.Fatalf("Satori 服务器运行时出错: %v", err)
			}
		}()

		// 使用通道来等待信号
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// 等待信号
		<-sigCh

		log.Info("正在关闭 Satori 服务器...")

		// 关闭 WebSocket 服务器
		if err := p.WebSocket.Close(); err != nil {
			log.Errorf("关闭 WebSocket 服务器时出错: %v", err)
		}

		// 使用一个超时关闭 http.Server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Errorf("关闭 Satori 服务器时出错: %v", err)
		}
	}
}

// ReadyHandler 自定义 ReadyHandler 感知连接成功事件
func ReadyHandler() event.ReadyHandler {
	return func(event *dto.WSPayload, data *dto.WSReadyData) {
		log.Infof("连接成功，欢迎使用 %s ！", data.User.Username)
		processor.SetStatus("qq", login.ONLINE)

		// 构建事件
		id, err := processor.HashEventID("READY-QQ" + time.Now().String())
		if err != nil {
			log.Errorf("构建事件 ID 时出错: %v", err)
			return
		}

		satoriEvent := &signaling.Event{
			Id:        id,
			Type:      signaling.EVENT_TYPE_LOGIN_ADDED,
			Platform:  "qq",
			SelfId:    processor.SelfId,
			Timestamp: time.Now().UnixNano() / 1e6,
			Login: &login.Login{
				User:     processor.GetBot("qq"),
				SelfId:   processor.SelfId,
				Platform: "qq",
				Status:   login.ONLINE,
			},
		}
		p.BroadcastEvent(satoriEvent)

		processor.SetStatus("qqguild", login.ONLINE)

		// 构建事件
		id, err = processor.HashEventID("READY-QQGUILD" + time.Now().String())
		if err != nil {
			log.Errorf("构建事件 ID 时出错: %v", err)
			return
		}

		satoriEvent = &signaling.Event{
			Id:        id,
			Type:      signaling.EVENT_TYPE_LOGIN_ADDED,
			Platform:  "qqguild",
			SelfId:    processor.SelfId,
			Timestamp: time.Now().UnixNano() / 1e6,
			Login: &login.Login{
				User:     processor.GetBot("qqguild"),
				SelfId:   processor.SelfId,
				Platform: "qqguild",
				Status:   login.ONLINE,
			},
		}
		p.BroadcastEvent(satoriEvent)
	}
}

// ErrorNotifyHandler 处理当 ws 链接发送错误的事件
func ErrorNotifyHandler() event.ErrorNotifyHandler {
	return func(err error) {
		log.Errorf("QQ 开放平台连接出现错误：%v", err)

		processor.SetStatus("qq", login.OFFLINE)

		// 构建事件
		id, err := processor.HashEventID("ERROR-QQ" + err.Error())
		if err != nil {
			log.Errorf("构建事件 ID 时出错: %v", err)
			return
		}

		satoriEvent := &signaling.Event{
			Id:        id,
			Type:      signaling.EVENT_TYPE_LOGIN_REMOVED,
			Platform:  "qq",
			SelfId:    processor.SelfId,
			Timestamp: time.Now().UnixNano() / 1e6,
			Login: &login.Login{
				User:     processor.GetBot("qq"),
				SelfId:   processor.SelfId,
				Platform: "qq",
				Status:   login.OFFLINE,
			},
		}
		p.BroadcastEvent(satoriEvent)

		processor.SetStatus("qqguild", login.OFFLINE)

		// 构建事件
		id, err = processor.HashEventID("ERROR-QQGUILD" + err.Error())
		if err != nil {
			log.Errorf("构建事件 ID 时出错: %v", err)
			return
		}

		satoriEvent = &signaling.Event{
			Id:        id,
			Type:      signaling.EVENT_TYPE_LOGIN_REMOVED,
			Platform:  "qqguild",
			SelfId:    processor.SelfId,
			Timestamp: time.Now().UnixNano() / 1e6,
			Login: &login.Login{
				User:     processor.GetBot("qqguild"),
				SelfId:   processor.SelfId,
				Platform: "qqguild",
				Status:   login.OFFLINE,
			},
		}
		p.BroadcastEvent(satoriEvent)
	}
}

// HelloHandler 处理 Hello 事件
func HelloHandler() event.HelloHandler {
	return func(event *dto.WSPayload) {
		processor.SetStatus("qq", login.ONLINE)

		// 构建事件
		id, err := processor.HashEventID("HELLO-QQ" + time.Now().String())
		if err != nil {
			log.Errorf("构建事件 ID 时出错: %v", err)
			return
		}

		satoriEvent := &signaling.Event{
			Id:        id,
			Type:      signaling.EVENT_TYPE_LOGIN_UPDATED,
			Platform:  "qq",
			SelfId:    processor.SelfId,
			Timestamp: time.Now().UnixNano() / 1e6,
			Login: &login.Login{
				User:     processor.GetBot("qq"),
				SelfId:   processor.SelfId,
				Platform: "qq",
				Status:   login.ONLINE,
			},
		}
		p.BroadcastEvent(satoriEvent)

		processor.SetStatus("qqguild", login.ONLINE)

		// 构建事件
		id, err = processor.HashEventID("HELLO-QQGUILD" + time.Now().String())
		if err != nil {
			log.Errorf("构建事件 ID 时出错: %v", err)
			return
		}

		satoriEvent = &signaling.Event{
			Id:        id,
			Type:      signaling.EVENT_TYPE_LOGIN_ADDED,
			Platform:  "qqguild",
			SelfId:    processor.SelfId,
			Timestamp: time.Now().UnixNano() / 1e6,
			Login: &login.Login{
				User:     processor.GetBot("qqguild"),
				SelfId:   processor.SelfId,
				Platform: "qqguild",
				Status:   login.ONLINE,
			},
		}
		p.BroadcastEvent(satoriEvent)
	}
}

// ReconnectHandler 处理 Reconnect 事件
func ReconnectHandler() event.ReconnectHandler {
	return func(event *dto.WSPayload) {
		processor.SetStatus("qq", login.RECONNECT)

		// 构建事件
		id, err := processor.HashEventID("RECONNECT-QQ" + time.Now().String())
		if err != nil {
			log.Errorf("构建事件 ID 时出错: %v", err)
			return
		}

		satoriEvent := &signaling.Event{
			Id:        id,
			Type:      signaling.EVENT_TYPE_LOGIN_UPDATED,
			Platform:  "qq",
			SelfId:    processor.SelfId,
			Timestamp: time.Now().UnixNano() / 1e6,
			Login: &login.Login{
				User:     processor.GetBot("qq"),
				SelfId:   processor.SelfId,
				Platform: "qq",
				Status:   login.RECONNECT,
			},
		}
		p.BroadcastEvent(satoriEvent)

		processor.SetStatus("qqguild", login.RECONNECT)

		// 构建事件
		id, err = processor.HashEventID("RECONNECT-QQGUILD" + time.Now().String())
		if err != nil {
			log.Errorf("构建事件 ID 时出错: %v", err)
			return
		}

		satoriEvent = &signaling.Event{
			Id:        id,
			Type:      signaling.EVENT_TYPE_LOGIN_UPDATED,
			Platform:  "qqguild",
			SelfId:    processor.SelfId,
			Timestamp: time.Now().UnixNano() / 1e6,
			Login: &login.Login{
				User:     processor.GetBot("qqguild"),
				SelfId:   processor.SelfId,
				Platform: "qqguild",
				Status:   login.RECONNECT,
			},
		}
		p.BroadcastEvent(satoriEvent)
	}
}

// PlainEventHandler 处理透传handler
func PlainEventHandler() event.PlainEventHandler {
	return func(event *dto.WSPayload, message []byte) error {
		// 默认为 qqguild
		return p.ProcessQQGuildInternal(event, message)
	}
}

// AudioEventHandler 音频机器人事件 handler
func AudioEventHandler() event.AudioEventHandler {
	return func(event *dto.WSPayload, data *dto.WSAudioData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// InteractionHandler 处理内联交互事件
func InteractionHandler() event.InteractionEventHandler {
	return func(event *dto.WSPayload, data *dto.WSInteractionData) error {
		return p.ProcessInteractionEvent(data)
	}
}

// ThreadEventHandler 处理论坛主题事件
func ThreadEventHandler() event.ThreadEventHandler {
	return func(event *dto.WSPayload, data *dto.WSThreadData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// PostEventHandler 处理论坛回帖事件
func PostEventHandler() event.PostEventHandler {
	return func(event *dto.WSPayload, data *dto.WSPostData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// ReplyEventHandler 处理论坛帖子回复事件
func ReplyEventHandler() event.ReplyEventHandler {
	return func(event *dto.WSPayload, data *dto.WSReplyData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// ForumAuditEventHandler 处理论坛帖子审核事件
func ForumAuditEventHandler() event.ForumAuditEventHandler {
	return func(event *dto.WSPayload, data *dto.WSForumAuditData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// GuildEventHandler 处理频道事件
func GuildEventHandler() event.GuildEventHandler {
	return func(event *dto.WSPayload, data *dto.WSGuildData) error {
		return p.ProcessGuildEvent(event, data)
	}
}

// MemberEventHandler 处理成员变更事件
func MemberEventHandler() event.GuildMemberEventHandler {
	return func(event *dto.WSPayload, data *dto.WSGuildMemberData) error {
		return p.ProcessMemberEvent(event, data)
	}
}

// ChannelEventHandler 处理子频道事件
func ChannelEventHandler() event.ChannelEventHandler {
	return func(event *dto.WSPayload, data *dto.WSChannelData) error {
		return p.ProcessChannelEvent(event, data)
	}
}

// CreateMessageHandler 处理消息事件 私域的事件 不at信息
func CreateMessageHandler() event.MessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSMessageData) error {
		return p.ProcessGuildNormalMessage(event, data)
	}
}

// ATMessageEventHandler 实现处理 频道at 消息的回调
func ATMessageEventHandler() event.ATMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSATMessageData) error {
		return p.ProcessGuildATMessage(event, data)
	}
}

// DirectMessageHandler 处理私信事件
func DirectMessageHandler() event.DirectMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSDirectMessageData) error {
		return p.ProcessChannelDirectMessage(event, data)
	}
}

// MessageDeleteEventHandler 处理私域消息删除事件
func MessageDeleteEventHandler() event.MessageDeleteEventHandler {
	return func(event *dto.WSPayload, data *dto.WSMessageDeleteData) error {
		return p.ProcessMessageDelete(event, data)
	}
}

// PublicMessageDeleteEventHandler 处理公域消息删除事件
func PublicMessageDeleteEventHandler() event.PublicMessageDeleteEventHandler {
	return func(event *dto.WSPayload, data *dto.WSPublicMessageDeleteData) error {
		return p.ProcessMessageDelete(event, data)
	}
}

// DirectMessageDeleteEventHandler 处理私聊消息删除事件
func DirectMessageDeleteEventHandler() event.DirectMessageDeleteEventHandler {
	return func(event *dto.WSPayload, data *dto.WSDirectMessageDeleteData) error {
		return p.ProcessMessageDelete(event, data)
	}
}

// MessageReactionEventHandler 处理表情表态事件
func MessageReactionEventHandler() event.MessageReactionEventHandler {
	return func(event *dto.WSPayload, data *dto.WSMessageReactionData) error {
		return p.ProcessMessageReaction(event, data)
	}
}

// MessageAuditEventHandler 处理消息审核事件
func MessageAuditEventHandler() event.MessageAuditEventHandler {
	return func(event *dto.WSPayload, data *dto.WSMessageAuditData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// GroupATMessageEventHandler 实现处理 群at 消息的回调
func GroupATMessageEventHandler() event.GroupATMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSGroupATMessageData) error {
		return p.ProcessGroupMessage(event, data)
	}
}

// GroupAddRobotEventHandler 实现处理 群添加机器人的回调
func GroupAddRobotEventHandler() event.GroupAddRobotEventHandler {
	return func(event *dto.WSPayload, data *dto.WSGroupAddRobotData) error {
		return p.ProcessGroupAddRobot(event, data)
	}
}

// GroupDelRobotEventHandler 实现处理 群删除机器人的回调
func GroupDelRobotEventHandler() event.GroupDelRobotEventHandler {
	return func(event *dto.WSPayload, data *dto.WSGroupDelRobotData) error {
		return p.ProcessGroupDelRobot(event, data)
	}
}

// C2CMessageEventHandler 实现处理私聊消息的回调
func C2CMessageEventHandler() event.C2CMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSC2CMessageData) error {
		return p.ProcessC2CMessage(event, data)
	}
}

func getHandlersByName(intentName string) ([]interface{}, bool) {
	switch intentName {
	case "DEFAULT": // 默认处理函数
		handlers := []interface{}{
			ReadyHandler(),
			ErrorNotifyHandler(),
			PlainEventHandler(),
		}
		return handlers, true
	case "GUILDS": // 频道事件
		handlers := []interface{}{
			GuildEventHandler(),
			ChannelEventHandler(),
		}
		return handlers, true
	case "GUILD_MEMBERS": // 频道成员事件
		handlers := []interface{}{MemberEventHandler()}
		return handlers, true
	case "GUILD_MESSAGES": // 私域频道消息事件
		handlers := []interface{}{
			CreateMessageHandler(),
			MessageDeleteEventHandler(),
		}
		return handlers, true
	case "GUILD_MESSAGE_REACTIONS": // 频道消息表情表态事件
		handlers := []interface{}{MessageReactionEventHandler()}
		return handlers, true
	case "DIRECT_MESSAGE": // 频道私信事件
		handlers := []interface{}{
			DirectMessageHandler(),
			DirectMessageDeleteEventHandler(),
		}
		return handlers, true
	case "OPEN_FORUMS_EVENT": // 公域论坛事件
		return nil, true
	case "AUDIO_OR_LIVE_CHANNEL_MEMBER": // 音频或直播频道成员事件
		return nil, true
	case "USER_MESSAGES": // 单聊/群聊消息事件
		handlers := []interface{}{
			GroupATMessageEventHandler(),
			GroupAddRobotEventHandler(),
			GroupDelRobotEventHandler(),
			C2CMessageEventHandler(),
		}
		return handlers, true
	case "INTERACTION": // 互动事件
		handlers := []interface{}{InteractionHandler()}
		return handlers, true
	case "MESSAGE_AUDIT": // 消息审核事件
		handlers := []interface{}{MessageAuditEventHandler()}
		return handlers, true
	case "FORUMS_EVENT": // 私域论坛事件
		handlers := []interface{}{
			ThreadEventHandler(),
			PostEventHandler(),
			ReplyEventHandler(),
			ForumAuditEventHandler(),
		}
		return handlers, true
	case "AUDIO_ACTION": // 音频机器人事件
		handlers := []interface{}{AudioEventHandler()}
		return handlers, true
	case "PUBLIC_GUILD_MESSAGES": // 公域频道消息事件
		handlers := []interface{}{
			ATMessageEventHandler(),
			PublicMessageDeleteEventHandler(),
		}
		return handlers, true
	default:
		log.Warnf("未知的 Intents : %s\n", intentName)
		return nil, false
	}
}
