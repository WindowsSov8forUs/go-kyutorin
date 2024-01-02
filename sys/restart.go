package sys

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
)

// RestartApplication 重启应用
func RestartApplication() {
	executableName, err := GetExecutableName()
	if err != nil {
		log.Fatalf("获取可执行文件名时出错: %v", err)
		os.Exit(1)
	}

	restarter := NewRestarter()
	if err := restarter.Restart(executableName); err != nil {
		log.Fatalf("重启应用时出错: %v", err)
		os.Exit(1)
	}

	if _, err := os.Create("restart.flag"); err != nil {
		log.Fatalf("创建重启标记文件时出错: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// GetExecutableName 获取可执行文件名
func GetExecutableName() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(executable, filepath.Ext(executable)), nil
}

// InitBase 解析参数并检测
func InitBase() {
	if runtime.GOOS == "windows" {
		if RunningByDoubleClick() {
			err := NoMoreDoubleClick()
			if err != nil {
				log.Errorf("遇到错误: %v", err)
				time.Sleep(time.Second * 5)
			}
			os.Exit(0)
		}
	} else {
		fmt.Printf("InitBase function is not implemented for %s\n", runtime.GOOS)
	}
}
