package global

import (
	"github.com/haierkeys/obsidian-better-sync-service/pkg/fileurl"
)

var (
	// 程序执行目录
	ROOT string
	Name string = "Obsidian Better Sync Service"
)

func init() {

	filename := fileurl.GetExePath()
	ROOT = filename + "/"

}
