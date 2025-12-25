package global

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
)

var (
	// 程序执行目录
	ROOT          string
	Name          string = "Fast Note Sync Service"
	WebClientName string = "Web"
)

func init() {

	filename := fileurl.GetExePath()
	ROOT = filename + "/"

}
