package processor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/satori-protocol-go/satori-model-go/pkg/login"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/operation"
)

type EventIDTable struct {
	m     sync.Map
	count int64
}

var table = &EventIDTable{}

// SaveEventID 保存事件 ID
func SaveEventID(id string) int64 {
	number := atomic.AddInt64(&table.count, 1) - 1
	table.m.Store(number, id)
	return number
}

// GetEventID 获取已经保存了的事件 ID
func GetEventID(id int64) string {
	if value, ok := table.m.Load(id); ok {
		return value.(string)
	}
	return ""
}

// BotMapping 机器人映射
type BotMapping struct {
	mapping map[string]*user.User
	mu      sync.Mutex
}

// StatusMapping 机器人状态映射
type StatusMapping struct {
	mapping map[string]login.LoginStatus
	mu      sync.Mutex
}

var SelfId string // 机器人 ID

var globalBotMapping = &BotMapping{
	mapping: make(map[string]*user.User),
}

var globalStatusMapping = &StatusMapping{
	mapping: make(map[string]login.LoginStatus),
}

// SetBot 设置机器人
func SetBot(platform string, bot *user.User) {
	globalBotMapping.mu.Lock()
	defer globalBotMapping.mu.Unlock()
	globalBotMapping.mapping[platform] = bot
}

// GetBot 获取机器人
func GetBot(platform string) *user.User {
	globalBotMapping.mu.Lock()
	defer globalBotMapping.mu.Unlock()
	return globalBotMapping.mapping[platform]
}

// GetBots 获取所有机器人
func GetBots() map[string]*user.User {
	globalBotMapping.mu.Lock()
	defer globalBotMapping.mu.Unlock()
	return globalBotMapping.mapping
}

// SetStatus 设置机器人状态
func SetStatus(platform string, status login.LoginStatus) {
	globalStatusMapping.mu.Lock()
	defer globalStatusMapping.mu.Unlock()
	globalStatusMapping.mapping[platform] = status
}

// GetStatus 获取机器人状态
func GetStatus(platform string) login.LoginStatus {
	globalStatusMapping.mu.Lock()
	defer globalStatusMapping.mu.Unlock()
	return globalStatusMapping.mapping[platform]
}

// GetReadyBody 创建 READY 信令的信令数据
func GetReadyBody() *operation.ReadyBody {
	var logins []*login.Login
	for platform, bot := range GetBots() {
		login := &login.Login{
			User:     bot,
			SelfId:   SelfId,
			Platform: platform,
			Status:   GetStatus(platform),
		}
		logins = append(logins, login)
	}
	return &operation.ReadyBody{
		Logins: logins,
	}
}

// DirectChannelIdMapping 私聊频道 ID 映射
type DirectChannelIdMapping struct {
	mapping map[string]string
	mu      sync.Mutex
}

// OpenIdMapping 开放 ID 映射
type OpenIdMapping struct {
	mapping map[string]string
	mu      sync.Mutex
}

// globalDirectChannelIdMapping 全局频道 ID 映射
var globalDirectChannelIdMappingInstance = &DirectChannelIdMapping{
	mapping: make(map[string]string),
}

// globalOpenIdMapping 全局开放 ID 映射
var globalOpenIdMappingInstance = &OpenIdMapping{
	mapping: make(map[string]string),
}

// GetDirectChannelGuild 获取私聊频道 ID
func GetDirectChannelGuild(channelId string) string {
	globalDirectChannelIdMappingInstance.mu.Lock()
	defer globalDirectChannelIdMappingInstance.mu.Unlock()
	return globalDirectChannelIdMappingInstance.mapping[channelId]
}

// SetDirectChannel 设置频道类型
func SetDirectChannel(channelId string, guildId string) {
	globalDirectChannelIdMappingInstance.mu.Lock()
	defer globalDirectChannelIdMappingInstance.mu.Unlock()
	globalDirectChannelIdMappingInstance.mapping[channelId] = guildId
}

// GetOpenIdType 获取开放 ID 类型
func GetOpenIdType(openId string) string {
	globalOpenIdMappingInstance.mu.Lock()
	defer globalOpenIdMappingInstance.mu.Unlock()
	return globalOpenIdMappingInstance.mapping[openId]
}

// SetOpenIdType 设置开放 ID 类型
func SetOpenIdType(openId string, openIdType string) {
	globalOpenIdMappingInstance.mu.Lock()
	defer globalOpenIdMappingInstance.mu.Unlock()
	globalOpenIdMappingInstance.mapping[openId] = openIdType
}

// DelOpenId 删除开放 ID
func DelOpenId(openId string) {
	globalOpenIdMappingInstance.mu.Lock()
	defer globalOpenIdMappingInstance.mu.Unlock()
	delete(globalOpenIdMappingInstance.mapping, openId)
}

// GetOpenIdData 获取开放 ID 数据
func GetOpenIdData() map[string]string {
	globalOpenIdMappingInstance.mu.Lock()
	defer globalOpenIdMappingInstance.mu.Unlock()
	return globalOpenIdMappingInstance.mapping
}

// Server 服务端接口
type Server interface {
	Run() error
	Send(*operation.Event)
	Close()
}

// Processor 消息处理器
type Processor struct {
	Api    openapi.OpenAPI
	ApiV2  openapi.OpenAPI
	Me     *dto.User
	Token  *token.Token
	Server Server
	conf   *config.Config
}

// NewProcessor 创建消息处理器
func NewProcessor(conf *config.Config) (*Processor, context.Context, error) {
	if conf.Account.Token == "" {
		return nil, nil, fmt.Errorf("bot account token is empty")
	}
	ctx := context.Background()

	// 获取 token
	token, err := getToken(conf, ctx)
	if err != nil {
		return nil, nil, err
	}

	// 创建 api
	api, apiV2, err := createOpenAPI(token, conf)
	if err != nil {
		return nil, nil, err
	}

	// 获取机器人信息
	me, err := getBotMe(api, ctx)
	if err != nil {
		return nil, nil, err
	}

	processor := &Processor{
		Api:    api,
		ApiV2:  apiV2,
		Me:     me,
		Token:  token,
		Server: nil,
		conf:   conf,
	}

	return processor, ctx, err
}

func (p *Processor) Run(ctx context.Context, server Server) error {
	p.Server = server

	err := establishWebSocket(p, p.ApiV2, p.Token, ctx, p.conf)
	if err != nil {
		return err
	}

	log.Info("已成功连接 QQ 开放平台")
	log.Infof("欢迎使用机器人：%s ！", p.Me.Username)

	// 启动 Satori 服务端
	go func() {
		if err := p.Server.Run(); err != nil {
			log.Fatalf("Satori 服务器运行时出错: %v", err)
		}
	}()

	return nil
}

// BroadcastEvent 向 Satori 应用发送事件
func (p *Processor) BroadcastEvent(event *operation.Event) error {
	p.Server.Send(event)
	return nil
}
