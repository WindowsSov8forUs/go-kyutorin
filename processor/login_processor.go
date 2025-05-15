package processor

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/satori-protocol-go/satori-model-go/pkg/login"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
	"github.com/tencent-connect/botgo/websocket"
)

var loginSnCounter int64 = 0

func generateLoginSn() int64 {
	return atomic.AddInt64(&loginSnCounter, 1)
}

// 构建登录事件 Login 资源
func buildLoginEventLogin(platform string) *login.Login {
	bot := GetBot(platform)
	return &login.Login{
		Sn:       generateLoginSn(),
		Platform: platform,
		User:     bot,
		Status:   GetStatus(platform),
		Adapter:  "kyutorin",
	}
}

// 构建非登录事件 Login 资源
func buildNonLoginEventLogin(platform string) *login.Login {
	bot := GetBot(platform)
	return &login.Login{
		Sn:       generateLoginSn(),
		Platform: platform,
		User:     bot,
	}
}

// getToken 获取 token
func getToken(conf *config.Config, ctx context.Context) (*token.Token, error) {
	// 获取 token
	token := token.BotToken(
		conf.Account.AppID,
		conf.Account.AppSecret,
		conf.Account.Token,
		token.TypeQQBot,
	)
	if err := token.InitToken(ctx); err != nil {
		return nil, err
	}
	return token, nil
}

// createOpenAPI 创建 openapi
func createOpenAPI(token *token.Token, conf *config.Config) (openapi.OpenAPI, openapi.OpenAPI, error) {
	var api openapi.OpenAPI
	var apiV2 openapi.OpenAPI

	if !conf.Account.Sandbox {
		// 创建 v1 版本 OpenAPI
		if err := botgo.SelectOpenAPIVersion(openapi.APIv1); err != nil {
			return nil, nil, err
		}
		api = botgo.NewOpenAPI(token).WithTimeout(10 * time.Second)

		// 创建 v2 版本 OpenAPI
		if err := botgo.SelectOpenAPIVersion(openapi.APIv2); err != nil {
			return nil, nil, err
		}
		apiV2 = botgo.NewOpenAPI(token).WithTimeout(10 * time.Second)
	} else {
		// 创建 v1 版本 OpenAPI
		if err := botgo.SelectOpenAPIVersion(openapi.APIv1); err != nil {
			return nil, nil, err
		}
		api = botgo.NewSandboxOpenAPI(token).WithTimeout(10 * time.Second)

		// 创建 v2 版本 OpenAPI
		if err := botgo.SelectOpenAPIVersion(openapi.APIv2); err != nil {
			return nil, nil, err
		}
		apiV2 = botgo.NewSandboxOpenAPI(token).WithTimeout(10 * time.Second)
	}

	return api, apiV2, nil
}

// getBotMe 获取机器人信息
func getBotMe(api openapi.OpenAPI, ctx context.Context, conf *config.Config) (*dto.User, error) {
	me, err := api.Me(ctx)
	if err != nil {
		return nil, err
	}
	qqBot := &user.User{
		Id:     strconv.Itoa(int(conf.Account.BotID)),
		Name:   me.Username,
		Avatar: me.Avatar,
		IsBot:  me.Bot,
	}
	qqGuildBot := &user.User{
		Id:     strconv.Itoa(int(conf.Account.AppID)),
		Name:   me.Username,
		Avatar: me.Avatar,
		IsBot:  me.Bot,
	}
	SetBot("qq", qqBot)
	SetBot("qqguild", qqGuildBot)
	SetStatus("qq", login.StatusOnline)
	SetStatus("qqguild", login.StatusOnline)
	SelfId = me.ID
	return me, nil
}

func establishWebSocket(p *Processor, apiV2 openapi.OpenAPI, token *token.Token, ctx context.Context, conf *config.Config) error {
	// 获取 WebSocket 信息
	wsInfo, err := apiV2.WS(ctx, nil, "")
	if err != nil {
		return err
	}

	// 定义和初始化 intent
	var intent dto.Intent = 0

	// 动态订阅 intent
	for _, intentName := range conf.Account.WebSocket.Intents {
		handlers, ok := p.getHandlersByName(intentName)
		if !ok {
			log.Warnf("未知的 intent : %s", intentName)
			continue
		}

		// 多次位与并订阅事件
		for _, handler := range handlers {
			intent |= websocket.RegisterHandlers(handler)
		}
	}

	log.Infof("订阅的 intent : %d", intent)

	// 启动 session manager 管理 websocket 连接
	// Gensokyo 强行设置分片数为 1 了，所以我也这么做吧
	go func() {
		wsInfo.Shards = conf.Account.WebSocket.Shards
		if err = botgo.NewSessionManager().Start(wsInfo, token, &intent); err != nil {
			log.Fatalf("启动 WebSocket 失败: %s", err)
		}
	}()
	return nil
}
