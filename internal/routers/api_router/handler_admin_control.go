package api_router

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
	"go.uber.org/zap"
)

// AdminControlHandler Admin control configuration API router handler
// AdminControlHandler 管理控制配置 API 路由处理器
// Uses App Container to inject dependencies
// 使用 App Container 注入依赖
type AdminControlHandler struct {
	*Handler
}

// NewAdminControlHandler creates AdminControlHandler instance
// NewAdminControlHandler 创建 AdminControlHandler 实例
func NewAdminControlHandler(a *app.App) *AdminControlHandler {
	return &AdminControlHandler{
		Handler: NewHandler(a),
	}
}

// webGUIConfig WebGUI configuration response structure (public interface)
// webGUIConfig WebGUI 配置响应结构（公开接口）
type webGUIConfig struct {
	FontSet          string `json:"fontSet"`          // Font set // 字体设置
	RegisterIsEnable bool   `json:"registerIsEnable"` // Registration enablement // 是否开启注册
	AdminUID         int    `json:"adminUid"`         // Admin UID // 管理员 UID
}

// adminConfig Admin configuration structure (admin interface)
// adminConfig 管理员配置结构（管理员接口）
type adminConfig struct {
	FontSet                 string `json:"fontSet" form:"fontSet"`                                           // Font set // 字体设置
	RegisterIsEnable        bool   `json:"registerIsEnable" form:"registerIsEnable"`                         // Registration enablement // 是否开启注册
	FileChunkSize           string `json:"fileChunkSize,omitempty" form:"fileChunkSize"`                     // File chunk size // 文件分块大小
	SoftDeleteRetentionTime string `json:"softDeleteRetentionTime,omitempty" form:"softDeleteRetentionTime"` // Soft delete retention time // 软删除保留时间
	UploadSessionTimeout    string `json:"uploadSessionTimeout,omitempty" form:"uploadSessionTimeout"`       // Upload session timeout // 上传会话超时时间
	HistoryKeepVersions     int    `json:"historyKeepVersions,omitempty" form:"historyKeepVersions"`         // History versions to keep // 历史版本保留数
	HistorySaveDelay        string `json:"historySaveDelay,omitempty" form:"historySaveDelay"`               // History save delay // 历史保存延迟
	DefaultAPIFolder        string `json:"defaultApiFolder,omitempty" form:"defaultApiFolder"`               // Default API folder // 默认 API 目录
	AdminUID                int    `json:"adminUid" form:"adminUid"`                                         // Admin UID // 管理员 UID
	AuthTokenKey            string `json:"authTokenKey" form:"authTokenKey"`                                 // Auth token key // 认证 Token 密钥
	TokenExpiry             string `json:"tokenExpiry" form:"tokenExpiry"`                                   // Token expiry // Token 有效期
	ShareTokenKey           string `json:"shareTokenKey" form:"shareTokenKey"`                               // Share token key // 分享 Token 密钥
	ShareTokenExpiry        string `json:"shareTokenExpiry" form:"shareTokenExpiry"`                         // Share token expiry // 分享 Token 有效期
}

// ngrokConfig Ngrok tunnel configuration
// ngrokConfig Ngrok 隧道配置
type ngrokConfig struct {
	Enabled   bool   `json:"enabled" form:"enabled"`     // Whether to enable ngrok tunnel // 是否启用 ngrok 隧道
	AuthToken string `json:"authToken" form:"authToken"` // ngrok auth token // ngrok 认证令牌
	Domain    string `json:"domain" form:"domain"`       // Custom domain // 自定义域名
}

// cloudflareConfig Cloudflare tunnel configuration
// cloudflareConfig Cloudflare 隧道配置
type cloudflareConfig struct {
	Enabled    bool   `json:"enabled" form:"enabled"`       // Whether to enable cloudflare tunnel // 是否启用 cloudflare 隧道
	Token      string `json:"token" form:"token"`           // cloudflare tunnel token // cloudflare 隧道令牌
	LogEnabled bool   `json:"logEnabled" form:"logEnabled"` // Whether to enable cloudflare tunnel logging // 是否开启 cloudflare 隧道日志
}

// SystemInfo system information response structure
// SystemInfo 系统信息响应结构
type SystemInfo struct {
	StartTime     time.Time   `json:"startTime"`     // Start time // 启动时间
	Uptime        float64     `json:"uptime"`        // Uptime (seconds) // 运行时间（秒）
	RuntimeStatus RuntimeInfo `json:"runtimeStatus"` // Go runtime status // Go 运行时状态
	CPU           CPUInfo     `json:"cpu"`           // CPU information // CPU 信息
	Memory        MemoryInfo  `json:"memory"`        // Memory information // 内存信息
	Host          HostInfo    `json:"host"`          // Host information // 主机信息
	Process       ProcessInfo `json:"process"`       // Process information // 进程信息
}

// CPUInfo CPU information
// CPUInfo CPU 信息
type CPUInfo struct {
	ModelName     string    `json:"modelName"`     // Model name // 型号
	PhysicalCores int       `json:"physicalCores"` // Physical cores // 物理核心数
	LogicalCores  int       `json:"logicalCores"`  // Logical cores // 逻辑核心数
	Percent       []float64 `json:"percent"`       // Usage percentage per core // 每个核心的使用率
	LoadAvg       *LoadInfo `json:"loadAvg"`       // Load average // 平均负载
}

// LoadInfo system load information
type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// MemoryInfo memory information
type MemoryInfo struct {
	Total           uint64  `json:"total"`           // Total physical memory // 系统总内存
	Available       uint64  `json:"available"`       // Available memory // 可用内存
	Used            uint64  `json:"used"`            // Used memory // 已用内存
	UsedPercent     float64 `json:"usedPercent"`     // Memory usage percentage // 内存使用率
	SwapTotal       uint64  `json:"swapTotal"`       // Total swap space // 交换区总量
	SwapUsed        uint64  `json:"swapUsed"`        // Used swap space // 交换区已用
	SwapUsedPercent float64 `json:"swapUsedPercent"` // Swap usage percentage // 交换区使用率
}

// HostInfo host identification information
type HostInfo struct {
	Hostname       string    `json:"hostname"`       // Hostname // 主机名
	OS             string    `json:"os"`             // Operating system // 操作系统
	OSPretty       string    `json:"osPretty"`       // Detailed OS name // 详细操作系统名称
	Platform       string    `json:"platform"`       // Platform name // 平台
	Arch           string    `json:"arch"`           // Architecture // 架构
	KernelVersion  string    `json:"kernelVersion"`  // Kernel version // 内核版本
	Uptime         uint64    `json:"uptime"`         // System uptime // 系统运行时间
	CurrentTime    time.Time `json:"currentTime"`    // Current system time // 当前系统时间
	TimeZone       string    `json:"timezone"`       // Time zone name // 时区名称
	TimeZoneOffset int       `json:"timezoneOffset"` // Time zone offset in seconds // 时区偏移（秒）
}

// ProcessInfo current process information
type ProcessInfo struct {
	PID           int32   `json:"pid"`           // Process ID
	PPID          int32   `json:"ppid"`          // Parent Process ID
	Name          string  `json:"name"`          // Process Name
	CPUPercent    float64 `json:"cpuPercent"`    // CPU Usage percentage
	MemoryPercent float32 `json:"memoryPercent"` // Memory Usage percentage
}

// RuntimeInfo Go runtime information
// RuntimeInfo Go 运行时信息
type RuntimeInfo struct {
	NumGoroutine int    `json:"numGoroutine"` // Number of goroutines // Goroutine 数量
	MemAlloc     uint64 `json:"memAlloc"`     // Allocated memory (bytes) // 已分配内存（字节）
	MemTotal     uint64 `json:"memTotal"`     // Total memory allocated (bytes) // 累计分配内存（字节）
	MemSys       uint64 `json:"memSys"`       // Memory obtained from system (bytes) // 从系统获取的内存（字节）
	HeapSys      uint64 `json:"heapSys"`      // Memory obtained from system for heap (bytes) // 堆占用的系统内存
	HeapIdle     uint64 `json:"heapIdle"`     // Memory in idle spans (bytes) // 空闲 Span 占用的内存
	HeapInuse    uint64 `json:"heapInuse"`    // Memory in in-use spans (bytes) // 正在使用的 Span 占用的内存
	HeapReleased uint64 `json:"heapReleased"` // Memory released to OS (bytes) // 释放回操作系统的内存（字节）
	StackSys     uint64 `json:"stackSys"`     // Memory obtained from system for stack (bytes) // 栈占用的系统内存
	MSpanSys     uint64 `json:"mSpanSys"`     // Memory obtained from system for mspan (bytes) // mspan 占用的系统内存
	MCacheSys    uint64 `json:"mCacheSys"`    // Memory obtained from system for mcache (bytes) // mcache 占用的系统内存
	BuckHashSys  uint64 `json:"buckHashSys"`  // Memory obtained from system for profiling bucket hash table (bytes) // 分析桶哈希表占用的系统内存
	GCSys        uint64 `json:"gcSys"`        // Memory obtained from system for metadata for GC (bytes) // GC 元数据占用的系统内存
	OtherSys     uint64 `json:"otherSys"`     // Other system memory (bytes) // 其他系统内存
	NextGC       uint64 `json:"nextGc"`       // Target heap size for the next GC cycle // 下次 GC 的目标堆大小
	NumGC        uint32 `json:"numGc"`        // Number of completed GC cycles // GC 次数
}

// Config retrieves WebGUI configuration (public interface)
// @Summary Get WebGUI basic config
// @Description Get non-sensitive configuration required for frontend display, such as font settings, registration status, etc.
// @Tags Config
// @Produce json
// @Success 200 {object} pkgapp.Res{data=webGUIConfig} "Success"
// @Router /api/webgui/config [get]
func (h *AdminControlHandler) Config(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	data := webGUIConfig{
		FontSet:          cfg.WebGUI.FontSet,
		RegisterIsEnable: cfg.User.RegisterIsEnable,
		AdminUID:         cfg.User.AdminUID,
	}
	response.ToResponse(code.Success.WithData(data))
}

// GetConfig retrieves admin configuration (requires admin privileges)
// @Summary Get full admin config
// @Description Get full system configuration information, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=adminConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config [get]
func (h *AdminControlHandler) GetConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.WebGUI.GetConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Deny access if AdminUID is configured and current user is not an admin
	// 当配置了管理员 UID 且当前用户不是管理员时，拒绝访问
	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	data := &adminConfig{
		FontSet:                 cfg.WebGUI.FontSet,
		RegisterIsEnable:        cfg.User.RegisterIsEnable,
		FileChunkSize:           cfg.App.FileChunkSize,
		SoftDeleteRetentionTime: cfg.App.SoftDeleteRetentionTime,
		UploadSessionTimeout:    cfg.App.UploadSessionTimeout,
		HistoryKeepVersions:     cfg.App.HistoryKeepVersions,
		HistorySaveDelay:        cfg.App.HistorySaveDelay,
		// DefaultAPIFolder:        cfg.App.DefaultAPIFolder,
		AdminUID:         cfg.User.AdminUID,
		AuthTokenKey:     cfg.Security.AuthTokenKey,
		TokenExpiry:      cfg.Security.TokenExpiry,
		ShareTokenKey:    cfg.Security.ShareTokenKey,
		ShareTokenExpiry: cfg.Security.ShareTokenExpiry,
	}

	response.ToResponse(code.Success.WithData(data))
}

// UpdateConfig updates admin configuration (requires admin privileges)
// @Summary Update admin config
// @Description Modify full system configuration information, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body adminConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res{data=adminConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config [post]
func (h *AdminControlHandler) UpdateConfig(c *gin.Context) {
	params := &adminConfig{}
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		logger.Error("apiRouter.WebGUI.UpdateConfig.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.WebGUI.UpdateConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Deny access if AdminUID is configured and current user is not an admin
	// 当配置了管理员 UID 且当前用户不是管理员时，拒绝访问
	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	// Validate historyKeepVersions cannot be less than 100
	// 验证 historyKeepVersions 不能小于 100
	if params.HistoryKeepVersions > 0 && params.HistoryKeepVersions < 100 {
		logger.Warn("apiRouter.WebGUI.UpdateConfig invalid historyKeepVersions",
			zap.Int("value", params.HistoryKeepVersions))
		response.ToResponse(code.ErrorInvalidParams.WithDetails("historyKeepVersions must be at least 100"))
		return
	}

	// Validate historySaveDelay cannot be less than 10 seconds
	// 验证 historySaveDelay 不能小于 10 秒
	if params.HistorySaveDelay != "" {
		delay, err := util.ParseDuration(params.HistorySaveDelay)
		if err != nil {
			logger.Warn("apiRouter.WebGUI.UpdateConfig invalid historySaveDelay format",
				zap.String("value", params.HistorySaveDelay))
			response.ToResponse(code.ErrorInvalidParams.WithDetails("historySaveDelay format invalid, e.g. 10s, 1m"))
			return
		}
		if delay < 10*time.Second {
			logger.Warn("apiRouter.WebGUI.UpdateConfig historySaveDelay too small",
				zap.String("value", params.HistorySaveDelay))
			response.ToResponse(code.ErrorInvalidParams.WithDetails("historySaveDelay must be at least 10s"))
			return
		}
	}

	// Update configuration
	// 更新配置
	cfg.WebGUI.FontSet = params.FontSet
	cfg.User.RegisterIsEnable = params.RegisterIsEnable
	cfg.App.FileChunkSize = params.FileChunkSize
	cfg.App.SoftDeleteRetentionTime = params.SoftDeleteRetentionTime
	cfg.App.UploadSessionTimeout = params.UploadSessionTimeout
	cfg.App.HistoryKeepVersions = params.HistoryKeepVersions
	cfg.App.HistorySaveDelay = params.HistorySaveDelay
	//cfg.App.DefaultAPIFolder = params.DefaultAPIFolder
	cfg.User.AdminUID = params.AdminUID
	cfg.Security.AuthTokenKey = params.AuthTokenKey
	cfg.Security.TokenExpiry = params.TokenExpiry
	cfg.Security.ShareTokenKey = params.ShareTokenKey
	cfg.Security.ShareTokenExpiry = params.ShareTokenExpiry

	// Save configuration to file
	// 保存配置到文件
	if err := cfg.Save(); err != nil {
		logger.Error("apiRouter.WebGUI.UpdateConfig.Save err", zap.Error(err))
		response.ToResponse(code.ErrorConfigSaveFailed)
		return
	}

	response.ToResponse(code.Success.WithData(params))
}

// GetNgrokConfig retrieves Ngrok tunnel configuration (requires admin privileges)
// @Summary Get Ngrok config
// @Description Get Ngrok tunnel configuration, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=ngrokConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/ngrok [get]
func (h *AdminControlHandler) GetNgrokConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.GetNgrokConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	data := &ngrokConfig{
		Enabled:   cfg.Ngrok.Enabled,
		AuthToken: cfg.Ngrok.AuthToken,
		Domain:    cfg.Ngrok.Domain,
	}

	response.ToResponse(code.Success.WithData(data))
}

// UpdateNgrokConfig updates Ngrok tunnel configuration (requires admin privileges)
// @Summary Update Ngrok config
// @Description Modify Ngrok tunnel configuration, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body ngrokConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res{data=ngrokConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/ngrok [post]
func (h *AdminControlHandler) UpdateNgrokConfig(c *gin.Context) {
	params := &ngrokConfig{}
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		logger.Error("apiRouter.AdminControl.UpdateNgrokConfig.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.UpdateNgrokConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	cfg.Ngrok.Enabled = params.Enabled
	cfg.Ngrok.AuthToken = params.AuthToken
	cfg.Ngrok.Domain = params.Domain

	if err := cfg.Save(); err != nil {
		logger.Error("apiRouter.AdminControl.UpdateNgrokConfig.Save err", zap.Error(err))
		response.ToResponse(code.ErrorConfigSaveFailed)
		return
	}

	response.ToResponse(code.Success.WithData(params))
}

// GetCloudflareConfig retrieves Cloudflare tunnel configuration (requires admin privileges)
// @Summary Get Cloudflare config
// @Description Get Cloudflare tunnel configuration, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=cloudflareConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/cloudflare [get]
func (h *AdminControlHandler) GetCloudflareConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.GetCloudflareConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	data := &cloudflareConfig{
		Enabled:    cfg.Cloudflare.Enabled,
		Token:      cfg.Cloudflare.Token,
		LogEnabled: cfg.Cloudflare.LogEnabled,
	}

	response.ToResponse(code.Success.WithData(data))
}

// UpdateCloudflareConfig updates Cloudflare tunnel configuration (requires admin privileges)
// @Summary Update Cloudflare config
// @Description Modify Cloudflare tunnel configuration, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body cloudflareConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res{data=cloudflareConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/cloudflare [post]
func (h *AdminControlHandler) UpdateCloudflareConfig(c *gin.Context) {
	params := &cloudflareConfig{}
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		logger.Error("apiRouter.AdminControl.UpdateCloudflareConfig.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.UpdateCloudflareConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	cfg.Cloudflare.Enabled = params.Enabled
	cfg.Cloudflare.Token = params.Token
	cfg.Cloudflare.LogEnabled = params.LogEnabled

	if err := cfg.Save(); err != nil {
		logger.Error("apiRouter.AdminControl.UpdateCloudflareConfig.Save err", zap.Error(err))
		response.ToResponse(code.ErrorConfigSaveFailed)
		return
	}

	response.ToResponse(code.Success.WithData(params))
}

// GetSystemInfo retrieves system and runtime information (requires admin privileges)
// @Summary Get system and runtime info
// @Description Get system information and Go runtime data, requires admin privileges
// @Tags System
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=SystemInfo} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/systeminfo [get]
func (h *AdminControlHandler) GetSystemInfo(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.WebGUI.GetSystemInfo err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	// Go Runtime
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// CPU
	cpuInfoList, _ := cpu.Info()
	cpuModel := ""
	if len(cpuInfoList) > 0 {
		cpuModel = cpuInfoList[0].ModelName
	}
	physCores, _ := cpu.Counts(false)
	logicCores, _ := cpu.Counts(true)
	cpuPercents, _ := cpu.Percent(time.Second, true)
	loadStat, _ := load.Avg()

	// Memory
	vMem, _ := mem.VirtualMemory()
	swapMem, _ := mem.SwapMemory()

	// Host
	hInfo, _ := host.Info()

	// Process
	p, _ := process.NewProcess(int32(os.Getpid()))
	pName, _ := p.Name()
	pPPid, _ := p.Ppid()
	pCPU, _ := p.CPUPercent()
	pMem, _ := p.MemoryPercent()

	data := SystemInfo{
		StartTime: h.App.StartTime,
		Uptime:    time.Since(h.App.StartTime).Seconds(),
		RuntimeStatus: RuntimeInfo{
			NumGoroutine: runtime.NumGoroutine(),
			MemAlloc:     m.Alloc,
			MemTotal:     m.TotalAlloc,
			MemSys:       m.Sys,
			HeapSys:      m.HeapSys,
			HeapIdle:     m.HeapIdle,
			HeapInuse:    m.HeapInuse,
			HeapReleased: m.HeapReleased,
			StackSys:     m.StackSys,
			MSpanSys:     m.MSpanSys,
			MCacheSys:    m.MCacheSys,
			BuckHashSys:  m.BuckHashSys,
			GCSys:        m.GCSys,
			OtherSys:     m.OtherSys,
			NextGC:       m.NextGC,
			NumGC:        m.NumGC,
		},
		CPU: CPUInfo{
			ModelName:     cpuModel,
			PhysicalCores: physCores,
			LogicalCores:  logicCores,
			Percent:       cpuPercents,
			LoadAvg: &LoadInfo{
				Load1:  loadStat.Load1,
				Load5:  loadStat.Load5,
				Load15: loadStat.Load15,
			},
		},
		Memory: MemoryInfo{
			Total:           vMem.Total,
			Available:       vMem.Available,
			Used:            vMem.Used,
			UsedPercent:     vMem.UsedPercent,
			SwapTotal:       swapMem.Total,
			SwapUsed:        swapMem.Used,
			SwapUsedPercent: swapMem.UsedPercent,
		},
		Host: HostInfo{
			Hostname:      hInfo.Hostname,
			OS:            hInfo.OS,
			OSPretty:      util.GetOSPrettyName(),
			Platform:      hInfo.Platform,
			Arch:          hInfo.KernelArch,
			KernelVersion: hInfo.KernelVersion,
			Uptime:        hInfo.Uptime,
			CurrentTime:   time.Now(),
			TimeZone:      time.Now().Location().String(),
			TimeZoneOffset: func() int {
				_, offset := time.Now().Zone()
				return offset
			}(),
		},
		Process: ProcessInfo{
			PID:           p.Pid,
			PPID:          pPPid,
			Name:          pName,
			CPUPercent:    pCPU,
			MemoryPercent: pMem,
		},
	}

	response.ToResponse(code.Success.WithData(data))
}

// Upgrade triggers server automatic upgrade
// @Summary Trigger server upgrade
// @Description Download latest version and restart server
// @Tags System
// @Produce json
// @Security UserAuthToken
// @Param version query string true "Version to upgrade (e.g. 2.0.10 or latest)"
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/admin/upgrade [get]
func (h *AdminControlHandler) Upgrade(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	uid := pkgapp.GetUID(c)

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	var upgradeReq dto.UpgradeRequest
	if ok, validErrs := pkgapp.BindAndValid(c, &upgradeReq); !ok {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(validErrs.Errors()...))
		return
	}

	checkInfo := h.App.CheckVersion("")
	version := ""

	if upgradeReq.Version == "latest" {
		if !checkInfo.VersionIsNew {
			response.ToResponse(code.Success.WithDetails("Current version is already up to date"))
			return
		}
		version = checkInfo.VersionNewName
	} else {
		version = upgradeReq.Version
	}

	versionRaw := strings.TrimPrefix(version, "v")

	// Determine download URL
	// 确定下载地址
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Example: fast-note-sync-service-2.0.10-linux-amd64.tar.gz
	fileName := fmt.Sprintf("fast-note-sync-service-%s-%s-%s.tar.gz", versionRaw, goos, goarch)
	downloadURL := ""
	if checkInfo.GithubAvailable {
		// GitHub releases/download/[tag]/[filename]
		// Based on user feedback: URL should NOT have 'v' in the tag part if the tag itself doesn't have it
		downloadURL = fmt.Sprintf("https://github.com/haierkeys/fast-note-sync-service/releases/download/%s/%s", versionRaw, fileName)
	} else {
		// CNB download URL format
		downloadURL = fmt.Sprintf("https://cnb.cool/haierkeys/fast-note-sync-service/-/releases/download/%s/%s", versionRaw, fileName)
	}

	h.App.Logger().Info("Starting upgrade download", zap.String("url", downloadURL), zap.String("version", versionRaw))

	// Prepare temp directory
	// 使用 storage/temp/upgrade 作为临时目录
	tempDir := filepath.Join("storage", "temp", "upgrade")
	_ = os.RemoveAll(tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		response.ToResponse(code.Failed.WithDetails("Failed to create temp directory: " + err.Error()))
		return
	}

	// Download
	tarPath := filepath.Join(tempDir, fileName)
	if err := h.downloadFile(downloadURL, tarPath); err != nil {
		response.ToResponse(code.Failed.WithDetails("Download failed: " + err.Error()))
		return
	}

	// Extract
	binaryName := "fast-note-sync-service"
	if goos == "windows" {
		binaryName += ".exe"
	}
	extractedBinaryPath := filepath.Join(tempDir, binaryName)

	if err := h.extractBinary(tarPath, tempDir, binaryName); err != nil {
		response.ToResponse(code.Failed.WithDetails("Extract failed: " + err.Error()))
		return
	}

	// Trigger upgrade in App
	h.App.TriggerUpgrade(extractedBinaryPath)

	response.ToResponse(code.Success.WithDetails("Upgrade triggered, server is restarting..."))
}

// Restart triggers server automatic restart
// @Summary Trigger server restart
// @Description Gracefully restart the server
// @Tags System
// @Produce json
// @Security UserAuthToken
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/admin/restart [get]
func (h *AdminControlHandler) Restart(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	uid := pkgapp.GetUID(c)

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	currentBinary, err := os.Executable()
	if err != nil {
		response.ToResponse(code.Failed.WithDetails("Failed to get current executable path: " + err.Error()))
		return
	}

	h.App.TriggerUpgrade(currentBinary)

	response.ToResponse(code.Success.WithDetails("Restart triggered, server is restarting..."))
}

// GC triggers manual garbage collection and releases memory to OS (requires admin privileges)
// GC 手动触发垃圾回收并释放内存给操作系统（需要管理员权限）
// @Summary Trigger manual GC
// @Description Manually run Go runtime GC and release memory to OS, requires admin privileges
// @Tags System
// @Produce json
// @Security UserAuthToken
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/gc [get]
func (h *AdminControlHandler) GC(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	startTime := time.Now()
	// Trigger GC // 触发 GC
	runtime.GC()
	// Release memory to OS // 释放内存给操作系统
	debug.FreeOSMemory()
	duration := time.Since(startTime)

	runtime.ReadMemStats(&mAfter)

	memReleased := int64(mBefore.Alloc) - int64(mAfter.Alloc)
	logger.Info("Manual GC completed",
		zap.Duration("duration", duration),
		zap.Uint64("allocBefore", mBefore.Alloc),
		zap.Uint64("allocAfter", mAfter.Alloc),
		zap.Int64("released", memReleased),
	)

	data := gin.H{
		"duration":    duration.String(),
		"allocBefore": mBefore.Alloc,
		"allocAfter":  mAfter.Alloc,
		"released":    memReleased,
	}

	response.ToResponse(code.Success.WithData(data).WithDetails("Manual GC completed successfully"))
}

func (h *AdminControlHandler) downloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (h *AdminControlHandler) extractBinary(tarPath string, destDir string, binaryName string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Check if it's the binary we're looking for
		// Often files in tar.gz are in a subdirectory or have different names
		// In alpha-release.yml: tar -czvf ... . (contents of build/platform dir)
		if filepath.Base(header.Name) == binaryName {
			target := filepath.Join(destDir, binaryName)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
			return nil
		}
	}

	return fmt.Errorf("binary %s not found in archive", binaryName)
}

// CloudflaredTunnelDownload triggers cloudflared binary download (requires admin privileges)
// @Summary Download cloudflared binary
// @Description Trigger the download of cloudflared binary for the current platform
// @Tags System
// @Security UserAuthToken
// @Produce json
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/admin/cloudflared_tunnel_download [get]
func (h *AdminControlHandler) CloudflaredTunnelDownload(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	uid := pkgapp.GetUID(c)

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	h.App.Logger().Info("Starting manual cloudflared binary download via API")

	path, err := h.App.CloudflareService.DownloadBinary()
	if err != nil {
		h.App.Logger().Error("Manual cloudflared download failed", zap.Error(err))
		// 返回详细的错误提示（包含 DownloadBinary 中构造的建议）
		response.ToResponse(code.ErrorCloudflaredDownloadFailed.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(gin.H{"path": path}).WithDetails("Cloudflared binary is ready"))
}
