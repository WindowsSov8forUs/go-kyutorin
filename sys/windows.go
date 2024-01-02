package sys

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// WindowsRestarter Windows 平台重启器
type WindowsRestarter struct{}

// NewRestarter 创建重启器
func NewRestarter() *WindowsRestarter {
	return &WindowsRestarter{}
}

// Restart 重启应用
func (r *WindowsRestarter) Restart(executablePath string) error {
	executableDir, executableName := filepath.Split(executablePath)

	scriptContent := "@echo off\n" +
		"pushd " + strconv.Quote(executableDir) + "\n" +
		"start \"\" " + strconv.Quote(executableName) + "\n" +
		"popd\n"

	scriptName := "restart.bat"
	if err := os.WriteFile(scriptName, []byte(scriptContent), 0755); err != nil {
		return err
	}

	cmd := exec.Command("cmd.exe", "/C", scriptName)

	if err := cmd.Start(); err != nil {
		return err
	}

	os.Exit(0)

	return nil
}
