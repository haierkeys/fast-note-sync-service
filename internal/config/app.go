package config

// AppSettings application settings
// AppSettings 应用设置
type AppSettings struct {
	// DefaultPageSize default page size
	// DefaultPageSize 默认页面大小
	DefaultPageSize int `yaml:"default-page-size" default:"10"`
	// MaxPageSize maximum page size
	// MaxPageSize 最大页面大小
	MaxPageSize int `yaml:"max-page-size" default:"100"`
	// DefaultContextTimeout default context timeout duration
	// DefaultContextTimeout 默认上下文超时时间
	DefaultContextTimeout int `yaml:"default-context-timeout" default:"60"`

	// TempPath upload temporary path
	// TempPath 上传临时路径
	TempPath string `yaml:"temp-path" default:"storage/temp"`
	// IsReturnSussess whether to return success info
	// IsReturnSussess 是否返回成功信息
	IsReturnSussess bool `yaml:"is-return-sussess" default:"false"`
	// SoftDeleteRetentionTime retention time for soft deleted notes
	// SoftDeleteRetentionTime 软删除笔记保留时间
	SoftDeleteRetentionTime string `yaml:"soft-delete-retention-time" default:"7d"`
	// SyncLogRetentionTime retention time for sync logs
	// SyncLogRetentionTime 同步日志保留时间
	SyncLogRetentionTime string `yaml:"sync-log-retention-time" default:"30d"`
	// HistoryKeepVersions number of historical versions to keep, default 100; yaml 显式 0 = 无限保留不清理，nil 才用默认 100
	// HistoryKeepVersions 历史记录保留版本数，默认 100；yaml 显式 0 = 无限保留不清理，nil 才用默认 100
	HistoryKeepVersions *int `yaml:"history-keep-versions" default:"100"`
	// HistorySaveDelay historical record save delay time, supports format: 10s (seconds), 1m (minutes), default 10s
	// HistorySaveDelay历史记录保存延迟时间，支持格式：10s（秒）、1m（分钟），默认 10s
	HistorySaveDelay string `yaml:"history-save-delay" default:"10s"`
	// UploadSessionTimeout file upload session timeout duration
	// UploadSessionTimeout 文件上传会话超时时间
	UploadSessionTimeout string `yaml:"upload-session-timeout" default:"1d"`
	// FileChunkSize file chunk size
	// FileChunkSize 文件分片大小
	FileChunkSize string `yaml:"file-chunk-size" default:"512KB"`
	// DownloadSessionTimeout file chunk download timeout duration
	// DownloadSessionTimeout 文件分片下载超时时间
	DownloadSessionTimeout string `yaml:"download-session-timeout" default:"1h"`

	// Worker Pool configurations
	// Worker Pool 配置
	WorkerPoolMaxWorkers int `yaml:"worker-pool-max-workers" default:"100"`
	WorkerPoolQueueSize  int `yaml:"worker-pool-queue-size" default:"1000"`

	// Write Queue configurations
	// Write Queue 配置
	WriteQueueCapacity int    `yaml:"write-queue-capacity" default:"1000"`
	WriteQueueTimeout  string `yaml:"write-queue-timeout" default:"30s"`
	WriteQueueIdleTime string `yaml:"write-queue-idle-time" default:"10m"`

	// WebSocket configurations
	// WebSocket 配置
	WebSocketReadMaxPayloadSize   string `yaml:"ws-read-max-payload-size" default:"128MB"`
	WebSocketWriteMaxPayloadSize  string `yaml:"ws-write-max-payload-size" default:"128MB"`
	WebSocketParallelEnabled      *bool  `yaml:"ws-parallel-enabled" default:"true"`
	WebSocketParallelGolimit      int    `yaml:"ws-parallel-golimit" default:"3"`
	WebSocketCheckUtf8Enabled     *bool  `yaml:"ws-check-utf8-enabled" default:"true"`
	WebSocketCompressionEnabled   *bool  `yaml:"ws-compression-enabled" default:"true"`
	WebSocketCompressionLevel     int    `yaml:"ws-compression-level" default:"1"`
	WebSocketCompressionThreshold int    `yaml:"ws-compression-threshold" default:"512"`
	// PullSource data pull source: auto | github | cnb
	// PullSource 数据拉取源：auto | github | cnb
	PullSource string `yaml:"pull-source" default:"auto"`
	// PullReleaseChannel update version channel: stable | beta
	// PullReleaseChannel 更新版本通道：stable（正式版） | beta（测试版）
	PullReleaseChannel string `yaml:"pull-release-channel" default:"stable"`

	// ShortLink configurations
	// 短链配置
	ShortLink ShortLinkConfig `yaml:"short-link"`

	FtsBleveEnabled  *bool `yaml:"fts-bleve-enabled" default:"true"`    // Bleve FTS enabled flag // 是否启用 Bleve 全文搜索（默认启用）
	FtsBleveStoreRaw *bool `yaml:"fts-bleve-store-raw" default:"false"` // Bleve FTS store raw content flag // Bleve 全文搜索是否存储原始文本（默认启用为方案 B，若设为 false 则为仅索引不存储的方案 A）
	SyncDownChunkNum int `yaml:"sync-down-chunk-num" default:"200"` // Serial download sync page chunk size // 串行下载同步的分块数量
	SyncUpChunkNum   int `yaml:"sync-up-chunk-num" default:"100"`  // Serial upload sync batch size // 串行上传同步的分包大小

	// PipelineWindowUp negotiated upload sliding-window size for pv>=2 connections; 0 disables
	// the window (stop-and-wait, same as pre-3.6.0 behavior — this is the runtime rollback
	// switch). Read sites clamp to [0,32]. MUST be *int, not int: LoadConfig runs defaults.Set
	// again after yaml.Unmarshal to fill empty fields, and a plain int can't distinguish an
	// explicit `pipeline-window-up: 0` from an unset field — the explicit 0 would be silently
	// overwritten back to the default 8, half-disabling the rollback switch. With *int, an
	// explicit yaml 0 becomes a non-nil pointer that defaults.Set leaves alone.
	// PipelineWindowUp pv>=2 连接协商的上行滑动窗口大小；0 表示禁用窗口（stop-and-wait，与 3.6.0 前
	// 行为一致——即运行时回滚开关）。读取处钳制到 [0,32]。必须用 *int 而非 int：LoadConfig 在
	// yaml.Unmarshal 之后会再次 defaults.Set 填充空字段，普通 int 无法区分显式 `pipeline-window-up: 0`
	// 与未写字段——显式 0 会被静默覆盖回默认 8，导致回滚开关半失效。改用 *int 后，yaml 显式 0
	// 反序列化为非 nil 指针，defaults.Set 不会覆盖。
	PipelineWindowUp *int `yaml:"pipeline-window-up" default:"8"`
	// PipelineWindowDown negotiated download sliding-window size for pv>=2 connections; 0 disables
	// the window (stop-and-wait, same as pre-3.6.0 behavior). Read sites clamp to [0,16].
	// *int for the same explicit-0-vs-unset reason as PipelineWindowUp.
	// PipelineWindowDown pv>=2 连接协商的下行滑动窗口大小；0 表示禁用窗口（stop-and-wait，与 3.6.0 前行为一致）。读取处钳制到 [0,16]。
	// 与 PipelineWindowUp 相同的「显式 0 vs 未写」原因，使用 *int。
	PipelineWindowDown *int `yaml:"pipeline-window-down" default:"4"`
}

// clampWindow clamps a pipeline window size to [0, max]; negative values are treated as 0
// (disabled / stop-and-wait).
// clampWindow 将流水线窗口大小钳制到 [0, max]；负值视为 0（禁用 / stop-and-wait）。
func clampWindow(v, max int) int {
	if v < 0 {
		return 0
	}
	if v > max {
		return max
	}
	return v
}

// PipelineWindowUpClamped returns PipelineWindowUp clamped to [0,32]. A nil pointer (config
// built without going through LoadConfig/defaults.Set) falls back to the default 8 defensively.
// PipelineWindowUpClamped 返回钳制到 [0,32] 的 PipelineWindowUp。nil 指针（未经
// LoadConfig/defaults.Set 构造的配置）防御性回退到默认值 8。
func (a AppSettings) PipelineWindowUpClamped() int {
	if a.PipelineWindowUp == nil {
		return 8
	}
	return clampWindow(*a.PipelineWindowUp, 32)
}

// PipelineWindowDownClamped returns PipelineWindowDown clamped to [0,16]. A nil pointer falls
// back to the default 4 defensively.
// PipelineWindowDownClamped 返回钳制到 [0,16] 的 PipelineWindowDown。nil 指针防御性回退到默认值 4。
func (a AppSettings) PipelineWindowDownClamped() int {
	if a.PipelineWindowDown == nil {
		return 4
	}
	return clampWindow(*a.PipelineWindowDown, 16)
}

