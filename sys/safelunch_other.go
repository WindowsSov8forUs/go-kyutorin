//go:build !windows
// +build !windows

package sys

// RunningByDoubleClick 是否通过双击运行
func RunningByDoubleClick() bool {
	return true
}

// NoMoreDoubleClick 提示不要双击运行，并生成启动脚本
func NoMoreDoubleClick() error {
	return nil
}

// toHighDPI tries to raise DPI awareness context to DPI_AWARENESS_CONTEXT_UNAWARE_GDISCALED
func toHighDPI() {}

// BoxW of Win32 API. Check https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-messageboxw for more detail.
func boxW() {}

// GetConsoleWindows retrieves the window handle used by the console associated with the calling process.
func getConsoleWindows() {
	return
}
