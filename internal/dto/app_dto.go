// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

// VersionDTO version information for API response
// VersionDTO 版本信息 API 响应对象
type VersionDTO struct {
	Version                          string `json:"version"`                          // Current version // 当前版本
	GitTag                           string `json:"gitTag"`                           // Git tag // Git 标签
	BuildTime                        string `json:"buildTime"`                        // Build time // 构建时间
	VersionIsNew                     bool   `json:"versionIsNew"`                     // Is there a new version // 是否有新版本
	VersionNewName                   string `json:"versionNewName"`                   // New version name // 新版本名称
	VersionNewLink                   string `json:"versionNewLink"`                   // New version download link // 新版本下载链接
	VersionNewChangelog              string `json:"versionNewChangelog"`              // New version changelog link // 新版本更新日志链接
	VersionNewChangelogContent       string `json:"versionNewChangelogContent"`       // New version changelog content // 新版本更新日志内容
	PluginVersionNewName             string `json:"pluginVersionNewName"`             // New plugin version name // 插件新版本名称
	PluginVersionNewLink             string `json:"pluginVersionNewLink"`             // New plugin version link // 插件新版本链接
	PluginVersionNewChangelog        string `json:"pluginVersionNewChangelog"`        // New plugin version changelog link // 插件新版本更新日志链接
	PluginVersionNewChangelogContent string `json:"pluginVersionNewChangelogContent"` // New plugin version changelog content // 插件新版本更新日志内容
}

// UpgradeRequest upgrade request parameters
// UpgradeRequest 升级请求参数
type UpgradeRequest struct {
	Version string `form:"version" binding:"required"` // Version to upgrade (e.g. 2.0.10 or latest) // 升级版本
}
