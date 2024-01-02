//go:build windows
// +build windows

// go:build windows
package sys

import (
	"os"
	"path/filepath"
	"unsafe"

	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"

	"golang.org/x/sys/windows"
)

// RunningByDoubleClick 是否通过双击运行
func RunningByDoubleClick() bool {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := kernel32.NewProc("GetConsoleProcessList")
	if proc != nil {
		var pids [2]uint32
		var maxCount uint32 = 2
		ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&pids)), uintptr(maxCount))
		if ret > 1 {
			return false
		}
	}
	return true
}

// NoMoreDoubleClick 提示不要双击运行，并生成启动脚本
func NoMoreDoubleClick() error {
	toHighDPI()
	r := boxW(getConsoleWindows(), "请勿通过双击直接运行本程序, 这将导致一些非预料的后果.\n请在shell中运行./gensokyo.exe\n点击确认将释出安全启动脚本，点击取消则关闭程序", "警告", 0x00000030|0x00000001)
	if r == 2 {
		return nil
	}
	r = boxW(0, "点击确认将覆盖 run.bat，点击取消则关闭程序", "警告", 0x00000030|0x00000001)
	if r == 2 {
		return nil
	}
	f, err := os.OpenFile("run.bat", os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return err
	}
	if err != nil {
		log.Errorf("打开 run.bat失败: %v", err)
		return nil
	}
	_ = f.Truncate(0)

	ex, _ := os.Executable()
	exPath := filepath.Base(ex)
	_, err = f.WriteString("%Created by go-satori-qq. DO NOT EDIT ME!%\nstart cmd /K \"" + exPath + "\"")
	if err != nil {
		log.Errorf("写入 run.bat失败: %v", err)
		return nil
	}
	f.Close()
	boxW(0, "已释出安全启动脚本，请双击 run.bat 启动", "提示", 0x00000000)
	return nil
}

// toHighDPI tries to raise DPI awareness context to DPI_AWARENESS_CONTEXT_UNAWARE_GDISCALED
func toHighDPI() {
	systemAware := ^uintptr(2) + 1
	unawareGDIScaled := ^uintptr(5) + 1
	u32 := windows.NewLazySystemDLL("user32.dll")
	proc := u32.NewProc("SetThreadDpiAwarenessContext")
	if proc.Find() != nil {
		return
	}
	for i := unawareGDIScaled; i <= systemAware; i++ {
		_, _, _ = u32.NewProc("SetThreadDpiAwarenessContext").Call(i)
	}
}

// BoxW of Win32 API. Check https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-messageboxw for more detail.
func boxW(hwnd uintptr, caption, title string, flags uint) int {
	captionPtr, _ := windows.UTF16PtrFromString(caption)
	titlePtr, _ := windows.UTF16PtrFromString(title)
	u32 := windows.NewLazySystemDLL("user32.dll")
	ret, _, _ := u32.NewProc("MessageBoxW").Call(
		hwnd,
		uintptr(unsafe.Pointer(captionPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags))

	return int(ret)
}

// GetConsoleWindows retrieves the window handle used by the console associated with the calling process.
func getConsoleWindows() (hWnd uintptr) {
	hWnd, _, _ = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleWindow").Call()
	return
}
