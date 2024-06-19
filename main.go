package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/fileserver"
	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/WindowsSov8forUs/go-kyutorin/server"
	"github.com/WindowsSov8forUs/go-kyutorin/sys"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo"
)

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

		fmt.Println("已生成默认配置文件 config.yml，请修改后重启程序")
		fmt.Println("按下任意键继续...")
		fmt.Scanln()
		os.Exit(0)
	}

	// 加载配置
	conf, err := config.LoadConfig("config.yml")
	if err != nil {
		fmt.Printf("加载配置文件时出错: %v", err)
		os.Exit(0)
		return
	}

	// 配置日志等级
	log.SetLogLevel(conf.LogLevel)

	// 设置 gin 运行模式
	if conf.DebugMode {
		fmt.Println("正在 Debug 模式下运行服务器！")
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 设置 logger
	logger := log.GetLogger()
	botgo.SetLogger(logger)

	// 如果配置并未设置
	if conf.Account.Token == "" {
		fmt.Println("检测到未完成机器人配置，请修改配置文件后重启程序")
		os.Exit(0)
		return
	}

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

	// 初始化消息处理器
	p, ctx, err := processor.NewProcessor(conf)
	if err != nil {
		log.Fatalf("建立与 QQ 开放平台连接时出错: %v", err)
	}
	// 创建 Satori 服务端
	server, err := server.NewServer(p.Api, p.ApiV2, conf)
	if err != nil {
		log.Fatalf("建立 Satori 服务端时出错: %v", err)
	}
	// 运行消息处理器
	err = p.Run(ctx, server)
	if err != nil {
		log.Fatalf("应用启动时出错: %v", err)
	}

	// 使用通道来等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigCh

	log.Info("正在关闭 Satori 服务端...")

	server.Close()
}
