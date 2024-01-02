package log

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
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
	FileLogger *log.Logger
	Mutex      sync.Mutex
	Level      LogLevel
}

// 内部 Logger 对象
var logger Logger

// 初始化 Logger 对象
func init() {
	logger.FileLogger = log.New(getLogFile(), "", 0)
	logger.Level = INFO
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

// getTimestamp 获取当前时间
func getTimestamp() string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s]", timestamp)
}

// getColor 获取当前日志等级颜色
func getColor(level LogLevel) *color.Color {
	switch level {
	case FATAL:
		return color.New(color.FgHiRed)
	case ERROR:
		return color.New(color.FgRed)
	case WARN:
		return color.New(color.FgYellow)
	case INFO:
		return color.New(color.FgWhite)
	case DEBUG:
		return color.New(color.FgHiYellow)
	case TRACE:
		return color.New(color.FgGreen)
	default:
		return color.New(color.FgWhite)
	}
}

// getLevelName 获取日志等级名称
func getLevelName(level LogLevel) string {
	switch level {
	case FATAL:
		return "FATAL"
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// SetLogLevel 设置 Logger 等级
func SetLogLevel(level LogLevel) {
	logger.Level = level
}

// GetLogger 获取 Logger 对象
func GetLogger() *Logger {
	return &logger
}

// Println 打印日志
func (logger *Logger) Println(level LogLevel, v ...interface{}) {
	// 判断日志等级
	if level > logger.Level {
		return
	}
	// 获取当前时间
	timestamp := getTimestamp()
	// 获取当前日志等级颜色
	color := getColor(level)
	// 获取日志内容
	content := fmt.Sprintf("%s [%s]: %s", timestamp, getLevelName(level), fmt.Sprintln(v...))
	// 写入日志
	logger.Mutex.Lock()
	defer logger.Mutex.Unlock()
	color.Printf("%s", content)
	logger.FileLogger.Printf(content)
}

// Fatal 致命错误
func Fatal(v ...interface{}) {
	logger.Println(FATAL, v...)
	os.Exit(1)
}

// Fatalf 致命错误
func Fatalf(format string, v ...interface{}) {
	logger.Println(FATAL, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// logger 实例的 Error 方法
func (logger *Logger) Error(v ...interface{}) {
	logger.Println(DEBUG, v...)
}

// Error 错误
func Error(v ...interface{}) {
	logger.Println(ERROR, v...)
}

// logger 实例的 Errorf 方法
func (logger *Logger) Errorf(format string, v ...interface{}) {
	logger.Println(DEBUG, fmt.Sprintf(format, v...))
}

// Errorf 错误
func Errorf(format string, v ...interface{}) {
	logger.Println(ERROR, fmt.Sprintf(format, v...))
}

// logger 实例的 Warn 方法
func (logger *Logger) Warn(v ...interface{}) {
	logger.Println(DEBUG, v...)
}

// Warn 警告
func Warn(v ...interface{}) {
	logger.Println(WARN, v...)
}

// logger 实例的 Warnf 方法
func (logger *Logger) Warnf(format string, v ...interface{}) {
	logger.Println(DEBUG, fmt.Sprintf(format, v...))
}

// Warnf 警告
func Warnf(format string, v ...interface{}) {
	logger.Println(WARN, fmt.Sprintf(format, v...))
}

// logger 实例的 Info 方法
func (logger *Logger) Info(v ...interface{}) {
	logger.Println(DEBUG, v...)
}

// Info 信息
func Info(v ...interface{}) {
	logger.Println(INFO, v...)
}

// logger 实例的 Infof 方法
func (logger *Logger) Infof(format string, v ...interface{}) {
	logger.Println(DEBUG, fmt.Sprintf(format, v...))
}

// Infof 信息
func Infof(format string, v ...interface{}) {
	logger.Println(INFO, fmt.Sprintf(format, v...))
}

// logger 实例的 Debug 方法
func (logger *Logger) Debug(v ...interface{}) {
	logger.Println(DEBUG, v...)
}

// Debug 调试
func Debug(v ...interface{}) {
	logger.Println(DEBUG, v...)
}

// logger 实例的 Debugf 方法
func (logger *Logger) Debugf(format string, v ...interface{}) {
	logger.Println(DEBUG, fmt.Sprintf(format, v...))
}

// Debugf 调试
func Debugf(format string, v ...interface{}) {
	logger.Println(DEBUG, fmt.Sprintf(format, v...))
}

// Trace 追踪
func Trace(v ...interface{}) {
	logger.Println(TRACE, v...)
}

// Tracef 追踪
func Tracef(format string, v ...interface{}) {
	logger.Println(TRACE, fmt.Sprintf(format, v...))
}

// Sync 未实现的方法
func (logger *Logger) Sync() error {
	// TODO: 这个方法是做什么的？目前只是用来实现接口
	return nil
}
