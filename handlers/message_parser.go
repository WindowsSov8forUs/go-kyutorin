package handlers

import (
	"sync"

	"github.com/dezhishen/satori-model-go/pkg/login"
	"github.com/dezhishen/satori-model-go/pkg/user"
)

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
