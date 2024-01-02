package echo

import "sync"

// globalDirectChannelIdMapping 全局私聊频道 ID 映射
type globalDirectChannelIdMapping struct {
	mapping map[string]string
	mu      sync.Mutex
}

// globalOpenIdMapping 全局开放 ID 映射
type globalOpenIdMapping struct {
	mapping map[string]string
	mu      sync.Mutex
}

// globalDirectChannelIdMapping 全局频道 ID 映射
var globalDirectChannelIdMappingInstance = &globalDirectChannelIdMapping{
	mapping: make(map[string]string),
}

// globalOpenIdMapping 全局开放 ID 映射
var globalOpenIdMappingInstance = &globalOpenIdMapping{
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
