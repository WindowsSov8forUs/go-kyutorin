package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger 日志等级
type LogLevel int

const (
	OFF   LogLevel = iota // 关闭所有日志记录
	FATAL                 // 致命错误
	ERROR                 // 错误
	WARN                  // 警告
	INFO                  // 信息
	DEBUG                 // 调试
	TRACE                 // 追踪
	ALL                   // 所有日志记录
)

// Logger 日志结构体
type Logger struct {
	*logrus.Logger
	Mutex sync.Mutex
	Level LogLevel
}

// 内部 Logger 对象
var logger Logger

// CustomFormatter 自定义格式化器
type CustomFormatter struct {
	TimestampFormat string
	ForceColors     bool
}

// Format 实现 logrus.Formatter 接口
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 时间戳颜色（浅蓝色）
	timestamp := HiCyan("[%s]", entry.Time.Format(f.TimestampFormat))

	// 日志级别颜色和名称
	var levelColor ColorFunc
	var levelName string
	switch entry.Level {
	case logrus.PanicLevel:
		levelColor = PanicColor // 深红色（与 Fatal 相同）
		levelName = "PANIC"
	case logrus.FatalLevel:
		levelColor = FatalColor // 深红色
		levelName = "FATAL"
	case logrus.ErrorLevel:
		levelColor = ErrorColor // 红色
		levelName = "ERROR"
	case logrus.WarnLevel:
		levelColor = WarnColor // 橙色（高亮黄色，比普通黄色更深）
		levelName = "WARN"
	case logrus.InfoLevel:
		levelColor = InfoColor // 蓝色
		levelName = "INFO"
	case logrus.DebugLevel:
		levelColor = DebugColor // 黄色（普通黄色，比高亮黄色浅）
		levelName = "DEBUG"
	case logrus.TraceLevel:
		levelColor = TraceColor // 绿色
		levelName = "TRACE"
	default:
		levelColor = White
		levelName = "UNKNOWN"
	}

	level := levelColor("[%s]", levelName)

	// 组合日志消息
	return []byte(fmt.Sprintf("%s %s: %s\n", timestamp, level, entry.Message)), nil
}

// 初始化 Logger 对象
func init() {
	logger.Logger = logrus.New()

	// 设置自定义日志格式
	logger.SetFormatter(&CustomFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})

	// 配置多输出（控制台和文件）
	logFile := getLogFile()
	if logFile != nil {
		logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
	} else {
		logger.SetOutput(os.Stdout)
	}

	logger.Level = INFO
	logger.SetLevel(logrus.InfoLevel)
}

// getLogFile 获取当前 .log 文件对象
func getLogFile() *os.File {
	// 创建 log 文件夹
	err := os.MkdirAll("log", os.ModePerm)
	if err != nil {
		return nil
	}
	// 获取当前时间
	timestamp := time.Now().Format("2006-01-02")
	// 获取 .log 文件名
	logFileName := fmt.Sprintf("%s.log", timestamp)
	// 获取 .log 文件对象
	file, err := os.OpenFile("log/"+logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return nil
	}
	return file
}

// convertLogLevel 转换自定义日志级别到 Logrus 级别
func convertLogLevel(level LogLevel) logrus.Level {
	switch level {
	case FATAL:
		return logrus.FatalLevel
	case ERROR:
		return logrus.ErrorLevel
	case WARN:
		return logrus.WarnLevel
	case INFO:
		return logrus.InfoLevel
	case DEBUG:
		return logrus.DebugLevel
	case TRACE:
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}

// SetLogLevel 设置 Logger 等级
func SetLogLevel(level LogLevel) {
	logger.Level = level
	logger.SetLevel(convertLogLevel(level))
}

// GetLogger 获取 Logger 对象
func GetLogger() *Logger {
	return &logger
}

// Println 打印日志
func (l *Logger) Println(level LogLevel, v ...interface{}) {
	// 判断日志等级
	if level > l.Level {
		return
	}

	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	logrusLevel := convertLogLevel(level)
	l.Logger.Log(logrusLevel, v...)
}

// Fatal 致命错误
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Fatalf 致命错误
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// logger 实例的 Error 方法
func (l *Logger) Error(v ...interface{}) {
	l.Logger.Error(v...)
}

// Error 错误
func Error(v ...interface{}) {
	logger.Error(v...)
}

// logger 实例的 Errorf 方法
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Logger.Errorf(format, v...)
}

// Errorf 错误
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// logger 实例的 Warn 方法
func (l *Logger) Warn(v ...interface{}) {
	l.Logger.Warn(v...)
}

// Warn 警告
func Warn(v ...interface{}) {
	logger.Warn(v...)
}

// logger 实例的 Warnf 方法
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Logger.Warnf(format, v...)
}

// Warnf 警告
func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

// logger 实例的 Info 方法
func (l *Logger) Info(v ...interface{}) {
	l.Logger.Info(v...)
}

// Info 信息
func Info(v ...interface{}) {
	logger.Info(v...)
}

// logger 实例的 Infof 方法
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Logger.Infof(format, v...)
}

// Infof 信息
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// logger 实例的 Debug 方法
func (l *Logger) Debug(v ...interface{}) {
	l.Logger.Debug(v...)
}

// Debug 调试
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// logger 实例的 Debugf 方法
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Logger.Debugf(format, v...)
}

// Debugf 调试
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Trace 追踪
func Trace(v ...interface{}) {
	logger.Trace(v...)
}

// Tracef 追踪
func Tracef(format string, v ...interface{}) {
	logger.Tracef(format, v...)
}

// Sync 同步日志缓冲区
func (l *Logger) Sync() error {
	// Logrus 会自动同步，但为了兼容性保留此方法
	return nil
}
