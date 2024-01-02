package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/sys"

	"gopkg.in/yaml.v3"
)

var (
	instance *Config
	mu       sync.Mutex
)

// Config 配置
type Config struct {
	LogLevel log.LogLevel `yaml:"log_level"` // 日志等级
	Account  Account      `yaml:"account"`   // QQ 机器人账号配置
	Satori   Satori       `yaml:"satori"`    // Satori 配置
}

// Account QQ 机器人账号配置
type Account struct {
	BotID     uint64    `yaml:"bot_id"`     // 机器人 QQ 号
	AppID     uint64    `yaml:"app_id"`     // 机器人 ID
	Token     string    `yaml:"token"`      // 机器人令牌
	AppSecret string    `yaml:"app_secret"` // 机器人密钥
	Sandbox   bool      `yaml:"sandbox"`    // 是否使用沙箱环境
	WebSocket WebSocket `yaml:"websocket"`  // WebSocket 配置
}

// WebSocket QQ 机器人 WebSocket 配置
type WebSocket struct {
	Shards  uint32   `yaml:"shards"`  // 分片数
	Intents []string `yaml:"intents"` // 事件订阅
}

// Satori Satori 配置
type Satori struct {
	Version uint8  `yaml:"version"` // Satori 版本，目前只有 1
	Path    string `yaml:"path"`    // Satori 部署路径，可以为空
	Token   string `yaml:"token"`   // 鉴权令牌
	Server  Server `yaml:"server"`  // 服务器配置
}

// Server 服务器配置
type Server struct {
	Host string `yaml:"host"` // 服务器地址
	Port uint16 `yaml:"port"` // 服务器端口
}

// GetSatoriToken 获取 Satori 鉴权令牌
func GetSatoriToken() string {
	return instance.Satori.Token
}

// LoadConfig 加载配置
func LoadConfig(path string) (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	// 如果已经加载过配置，直接返回
	if instance != nil {
		return instance, nil
	}

	configData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err = yaml.Unmarshal(configData, config); err != nil {
		return nil, err
	}

	// 确保配置完整性
	if err = ensureConfigComplete(config, path); err != nil {
		return nil, err
	}

	instance = config
	return instance, nil
}

// ensureConfigComplete 检查配置是否完整
func ensureConfigComplete(config *Config, path string) error {
	// 读取配置文件
	configData, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 解析到结构体中
	currentConfig := &Config{}
	if err = yaml.Unmarshal(configData, currentConfig); err != nil {
		return err
	}

	// 解析默认配置模板
	defaultConfig := &Config{}
	if err = yaml.Unmarshal([]byte(ConfigTemplate), defaultConfig); err != nil {
		return err
	}

	// 使用反射找出缺失设置
	missingSettings, err := getMissingSettingsByReflection(currentConfig, defaultConfig)
	if err != nil {
		return err
	}

	// 使用文本比对找出缺失设置
	missingSettingsByText, err := getMissingSettingsByText(ConfigTemplate, string(configData))
	if err != nil {
		return err
	}

	// 合并缺失设置
	missingSettings = mergeMissingSettings(missingSettings, missingSettingsByText)

	// 如果有缺失设置，处理缺失配置行
	if len(missingSettings) > 0 {
		fmt.Printf("检测到配置文件不完整，缺失以下设置：\n%s", missingSettings)
		missingConfigLines, err := extractMissingConfigLines(missingSettings, ConfigTemplate)
		if err != nil {
			return err
		}

		// 追加到配置文件
		if err = appendToConfigFile(path, missingConfigLines); err != nil {
			return err
		}

		fmt.Printf("配置文件已更新，正在重新启动程序。")
		sys.RestartApplication()
	}

	return nil
}

// getMissingSettingsByReflection 使用反射来对比结构体并找出缺失的设置
func getMissingSettingsByReflection(currentConfig, defaultConfig *Config) (map[string]string, error) {
	missingSettings := make(map[string]string)
	currentVal := reflect.ValueOf(currentConfig).Elem()
	defaultVal := reflect.ValueOf(defaultConfig).Elem()

	for i := 0; i < currentVal.NumField(); i++ {
		field := currentVal.Type().Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || field.Type.Kind() == reflect.Int || field.Type.Kind() == reflect.Bool {
			continue // 跳过没有yaml标签的字段，或者字段类型为int或bool
		}
		yamlKeyName := strings.SplitN(yamlTag, ",", 2)[0]
		if isZeroOfUnderlyingType(currentVal.Field(i).Interface()) && !isZeroOfUnderlyingType(defaultVal.Field(i).Interface()) {
			missingSettings[yamlKeyName] = "missing"
		}
	}

	return missingSettings, nil
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// getMissingSettingsByText compares settings in two strings line by line, looking for missing keys.
func getMissingSettingsByText(templateContent, currentConfigContent string) (map[string]string, error) {
	templateKeys := extractKeysFromString(templateContent)
	currentKeys := extractKeysFromString(currentConfigContent)

	missingSettings := make(map[string]string)
	for key := range templateKeys {
		if _, found := currentKeys[key]; !found {
			missingSettings[key] = "missing"
		}
	}

	return missingSettings, nil
}

// extractKeysFromString reads a string and extracts the keys (text before the colon).
func extractKeysFromString(content string) map[string]bool {
	keys := make(map[string]bool)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			key := strings.TrimSpace(strings.Split(line, ":")[0])
			keys[key] = true
		}
	}
	return keys
}

// mergeMissingSettings 合并由反射和文本比对找到的缺失设置
func mergeMissingSettings(reflectionSettings, textSettings map[string]string) map[string]string {
	for k, v := range textSettings {
		reflectionSettings[k] = v
	}
	return reflectionSettings
}

func extractMissingConfigLines(missingSettings map[string]string, configTemplate string) ([]string, error) {
	var missingConfigLines []string

	lines := strings.Split(configTemplate, "\n")
	for yamlKey := range missingSettings {
		found := false
		// Create a regex to match the line with optional spaces around the colon
		regexPattern := fmt.Sprintf(`^\s*%s\s*:\s*`, regexp.QuoteMeta(yamlKey))
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %s", err)
		}

		for _, line := range lines {
			if regex.MatchString(line) {
				missingConfigLines = append(missingConfigLines, line)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("missing configuration for key: %s", yamlKey)
		}
	}

	return missingConfigLines, nil
}

func appendToConfigFile(path string, lines []string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("打开文件错误:", err)
		return err
	}
	defer file.Close()

	// 写入缺失的配置项
	for _, line := range lines {
		if _, err := file.WriteString("\n" + line); err != nil {
			fmt.Println("写入配置错误:", err)
			return err
		}
	}

	// 输出写入状态
	fmt.Println("配置已更新，写入到文件:", path)

	return nil
}
