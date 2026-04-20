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
	// LogSavePath log save path
	// LogSavePath 日志保存路径
	LogSavePath string `yaml:"log-save-fileurl"`
	// LogFile log filename
	// LogFile 日志文件名
	LogFile string `yaml:"log-file"`
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
	// HistoryKeepVersions number of historical versions to keep, default 100
	// HistoryKeepVersions 历史记录保留版本数，默认 100
	HistoryKeepVersions int `yaml:"history-keep-versions" default:"100"`
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
	WebSocketParallelEnabled      bool   `yaml:"ws-parallel-enabled" default:"true"`
	WebSocketParallelGolimit      int    `yaml:"ws-parallel-golimit" default:"8"`
	WebSocketCheckUtf8Enabled     bool   `yaml:"ws-check-utf8-enabled" default:"true"`
	WebSocketCompressionEnabled   bool   `yaml:"ws-compression-enabled" default:"true"`
	WebSocketCompressionLevel     int    `yaml:"ws-compression-level" default:"1"`
	WebSocketCompressionThreshold int    `yaml:"ws-compression-threshold" default:"512"`
	// PullSource data pull source: auto | github | cnb
	// PullSource 数据拉取源：auto | github | cnb
	PullSource string `yaml:"pull-source" default:"auto"`

	// ShortLink configurations
	// 短链配置
	ShortLink ShortLinkConfig `yaml:"short-link"`
}
