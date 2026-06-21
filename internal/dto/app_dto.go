// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"

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
	VersionNewChangelogContent       string                     `json:"versionNewChangelogContent"`       // New version changelog content // 新版本更新日志内容
	VersionHistory                   []pkgapp.HistoricalVersion `json:"versionHistory"`                   // Service version history // 服务端历史版本
	PluginVersionNewName             string                     `json:"pluginVersionNewName"`             // New plugin version name // 插件新版本名称
	PluginVersionNewLink             string                     `json:"pluginVersionNewLink"`             // New plugin version link // 插件新版本链接
	PluginVersionNewChangelog        string                     `json:"pluginVersionNewChangelog"`        // New plugin version changelog link // 插件新版本更新日志链接
	PluginVersionNewChangelogContent string                     `json:"pluginVersionNewChangelogContent"` // New plugin version changelog content // 插件新版本更新日志内容
	PluginVersionHistory             []pkgapp.HistoricalVersion `json:"pluginVersionHistory"`             // Plugin version history // 插件历史版本
}

// UpgradeRequest upgrade request parameters
// UpgradeRequest 升级请求参数
type UpgradeRequest struct {
	Version string `form:"version" binding:"required"` // Version to upgrade (e.g. 2.0.10 or latest) // 升级版本
}

// SourceProbeItem is one source's reachability + latency result.
// SourceProbeItem 单个源的可达性与延迟结果。
type SourceProbeItem struct {
	OK        bool  `json:"ok"`        // Whether the source is reachable // 该源是否可达
	LatencyMs int64 `json:"latencyMs"` // Round-trip latency in ms // 往返延迟（毫秒）
}

// SourceProbeDTO is the payload returned by GET /api/version/probe, consumed by
// the webgui "test latency" panel. Recommended is "github" or "cnb"; SelectedMode
// is the current configured pull-source mode (auto|github|cnb).
// SourceProbeDTO 是 GET /api/version/probe 的响应，供 webgui「测试延迟」面板使用。
// Recommended 为推荐的源（github 或 cnb）；SelectedMode 为当前配置的选源模式。
type SourceProbeDTO struct {
	GitHub       SourceProbeItem `json:"github"`
	CNB          SourceProbeItem `json:"cnb"`
	Recommended  string          `json:"recommended"`  // "github" or "cnb" // 推荐源
	SelectedMode string          `json:"selectedMode"` // current configured mode // 当前配置的模式
}
