package version

import "github.com/WindowsSov8forUs/glyccat/log"

// Version 版本号
var Version string = "Unknown"

var logo string = ""

func Logo() string {
	if logo == "" {
		logo += log.Green("\n   ██████╗ ██╗  ██╗   ██╗ ██████╗ ██████╗ █████╗ ████████╗\n")
		logo += log.Green("  ██╔════╝ ██║  ╚██╗ ██╔╝██╔════╝██╔════╝██╔══██╗╚══██╔══╝\n")
		logo += log.Yellow("  ██║  ███╗██║   ╚████╔╝ ██║     ██║     ███████║   ██║   \n")
		logo += log.Yellow("  ██║   ██║██║    ╚██╔╝  ██║     ██║     ██╔══██║   ██║   \n")
		logo += log.Cyan("  ╚██████╔╝███████╗██║   ╚██████╗╚██████╗██║  ██║   ██║   \n")
		logo += log.Cyan("   ╚═════╝ ╚══════╝╚═╝    ╚═════╝ ╚═════╝╚═╝  ╚═╝   ╚═╝   \n")
	}
	return logo
}
