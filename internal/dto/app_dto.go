// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

// VersionDTO version information for API response
// VersionDTO 版本信息 API 响应对象
type VersionDTO struct {
	Version        string `json:"version"`        // Current version // 当前版本
	GitTag         string `json:"gitTag"`         // Git tag // Git 标签
	BuildTime      string `json:"buildTime"`      // Build time // 构建时间
	VersionIsNew   bool   `json:"versionIsNew"`   // Is there a new version // 是否有新版本
	VersionNewName string `json:"versionNewName"` // New version name // 新版本名称
	VersionNewLink string `json:"versionNewLink"` // New version download link // 新版本下载链接
}
