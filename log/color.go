package log

import (
	"fmt"

	"github.com/fatih/color"
)

// 定义颜色函数类型
type ColorFunc func(format string, a ...interface{}) string

// 基础颜色函数
var (
	// 文本颜色
	Red     ColorFunc = color.New(color.FgRed).Sprintf
	Green   ColorFunc = color.New(color.FgGreen).Sprintf
	Yellow  ColorFunc = color.New(color.FgYellow).Sprintf
	Blue    ColorFunc = color.New(color.FgBlue).Sprintf
	Magenta ColorFunc = color.New(color.FgMagenta).Sprintf
	Cyan    ColorFunc = color.New(color.FgCyan).Sprintf
	White   ColorFunc = color.New(color.FgWhite).Sprintf
	Black   ColorFunc = color.New(color.FgBlack).Sprintf

	// 高亮颜色
	HiRed     ColorFunc = color.New(color.FgHiRed).Sprintf
	HiGreen   ColorFunc = color.New(color.FgHiGreen).Sprintf
	HiYellow  ColorFunc = color.New(color.FgHiYellow).Sprintf
	HiBlue    ColorFunc = color.New(color.FgHiBlue).Sprintf
	HiMagenta ColorFunc = color.New(color.FgHiMagenta).Sprintf
	HiCyan    ColorFunc = color.New(color.FgHiCyan).Sprintf
	HiWhite   ColorFunc = color.New(color.FgHiWhite).Sprintf
	HiBlack   ColorFunc = color.New(color.FgHiBlack).Sprintf

	// 加粗颜色
	BoldRed     ColorFunc = color.New(color.FgRed, color.Bold).Sprintf
	BoldGreen   ColorFunc = color.New(color.FgGreen, color.Bold).Sprintf
	BoldYellow  ColorFunc = color.New(color.FgYellow, color.Bold).Sprintf
	BoldBlue    ColorFunc = color.New(color.FgBlue, color.Bold).Sprintf
	BoldMagenta ColorFunc = color.New(color.FgMagenta, color.Bold).Sprintf
	BoldCyan    ColorFunc = color.New(color.FgCyan, color.Bold).Sprintf
	BoldWhite   ColorFunc = color.New(color.FgWhite, color.Bold).Sprintf

	// 日志级别专用颜色
	FatalColor ColorFunc = color.New(color.FgHiRed, color.Bold).Sprintf // 深红色加粗
	PanicColor ColorFunc = color.New(color.FgHiRed, color.Bold).Sprintf // 深红色加粗
	ErrorColor ColorFunc = color.New(color.FgRed).Sprintf               // 红色
	WarnColor  ColorFunc = color.New(color.FgYellow).Sprintf            // 橙色（黄色）
	InfoColor  ColorFunc = color.New(color.FgBlue).Sprintf              // 蓝色
	DebugColor ColorFunc = color.New(color.FgHiYellow).Sprintf          // 高亮黄色
	TraceColor ColorFunc = color.New(color.FgGreen).Sprintf             // 绿色
	TimeColor  ColorFunc = color.New(color.FgHiCyan).Sprintf            // 时间戳颜色（浅蓝色）
)

// 便捷函数 - 直接输出彩色文本（不换行）
func PrintRed(text string) {
	fmt.Print(Red("%s", text))
}

func PrintGreen(text string) {
	fmt.Print(Green("%s", text))
}

func PrintYellow(text string) {
	fmt.Print(Yellow("%s", text))
}

func PrintBlue(text string) {
	fmt.Print(Blue("%s", text))
}

func PrintCyan(text string) {
	fmt.Print(Cyan("%s", text))
}

// 便捷函数 - 直接输出彩色文本（带换行）
func PrintlnRed(text string) {
	fmt.Println(Red("%s", text))
}

func PrintlnGreen(text string) {
	fmt.Println(Green("%s", text))
}

func PrintlnYellow(text string) {
	fmt.Println(Yellow("%s", text))
}

func PrintlnBlue(text string) {
	fmt.Println(Blue("%s", text))
}

func PrintlnCyan(text string) {
	fmt.Println(Cyan("%s", text))
}

// 格式化输出函数
func PrintfRed(format string, a ...interface{}) {
	fmt.Print(Red(format, a...))
}

func PrintfGreen(format string, a ...interface{}) {
	fmt.Print(Green(format, a...))
}

func PrintfYellow(format string, a ...interface{}) {
	fmt.Print(Yellow(format, a...))
}

func PrintfBlue(format string, a ...interface{}) {
	fmt.Print(Blue(format, a...))
}

func PrintfCyan(format string, a ...interface{}) {
	fmt.Print(Cyan(format, a...))
}

// 获取日志级别对应的颜色函数
func GetLevelColorFunc(level string) ColorFunc {
	switch level {
	case "FATAL":
		return FatalColor
	case "PANIC":
		return PanicColor
	case "ERROR":
		return ErrorColor
	case "WARN":
		return WarnColor
	case "INFO":
		return InfoColor
	case "DEBUG":
		return DebugColor
	case "TRACE":
		return TraceColor
	default:
		return White
	}
}

// 创建自定义颜色函数
func NewColorFunc(attrs ...color.Attribute) ColorFunc {
	return color.New(attrs...).Sprintf
}

// 输出符号
var (
	SuccessMark = Green("✓")  // 成功
	FailMark    = Red("✗")    // 失败
	WarningMark = Yellow("⚠") // 警告
	InfoMark    = Cyan("ℹ")   // 信息
)
