package api_router

import (
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
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

// WebGUIHandler WebGUI configuration API router handler
// WebGUIHandler WebGUI 配置 API 路由处理器
// Uses App Container to inject dependencies
// 使用 App Container 注入依赖
type WebGUIHandler struct {
	*Handler
}

// NewWebGUIHandler creates WebGUIHandler instance
// NewWebGUIHandler 创建 WebGUIHandler 实例
func NewWebGUIHandler(a *app.App) *WebGUIHandler {
	return &WebGUIHandler{
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

// webGUIAdminConfig WebGUI administrator configuration structure (admin interface)
// webGUIAdminConfig WebGUI 管理员配置结构（管理员接口）
type webGUIAdminConfig struct {
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
	NumGC        uint32 `json:"numGc"`        // Number of completed GC cycles // GC 次数
}

// Config retrieves WebGUI configuration (public interface)
// @Summary Get WebGUI basic config
// @Description Get non-sensitive configuration required for frontend display, such as font settings, registration status, etc.
// @Tags Config
// @Produce json
// @Success 200 {object} pkgapp.Res{data=webGUIConfig} "Success"
// @Router /api/webgui/config [get]
func (h *WebGUIHandler) Config(c *gin.Context) {
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
// @Success 200 {object} pkgapp.Res{data=webGUIAdminConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config [get]
func (h *WebGUIHandler) GetConfig(c *gin.Context) {
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

	data := &webGUIAdminConfig{
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
// @Param params body webGUIAdminConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res{data=webGUIAdminConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config [post]
func (h *WebGUIHandler) UpdateConfig(c *gin.Context) {
	params := &webGUIAdminConfig{}
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
func (h *WebGUIHandler) GetSystemInfo(c *gin.Context) {
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
