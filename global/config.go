package global

import (
	"os"

	"github.com/haierkeys/obsidian-better-sync-service/pkg/fileurl"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/storage/local_fs"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	Config *config
)

type config struct {
	File     string
	Server   server          `yaml:"server"`
	Log      LogConfig       `yaml:"log"`
	Database Database        `yaml:"database"`
	App      app             `yaml:"app"`
	Security security        `yaml:"security"`
	LocalFS  local_fs.Config `yaml:"local-fs"`
}

type LogConfig struct {
	// Level, See also zapcore.ParseLevel.
	Level string `yaml:"level"`

	// File that logger will be writen into.
	// Default is stderr.
	File string `yaml:"file"`

	// Production enables json output.
	Production bool `yaml:"production"`
}

// Server is a struct that holds the server settings
type server struct {
	// RunMode is a string that holds the run mode of the server
	RunMode string `yaml:"run-mode"`
	// HttpPort is a string that holds the http port of the server
	HttpPort string `yaml:"http-port"`
	// ReadTimeout is a duration that holds the read timeout of the server
	ReadTimeout int `yaml:"read-timeout"`
	// WriteTimeout is a duration that holds the write timeout of the server
	WriteTimeout int `yaml:"write-timeout"`
	// PrivateHttpListen is a string that holds the private http listen address of the server
	PrivateHttpListen string `yaml:"private-http-listen"`
}

type security struct {
	AuthToken    string `yaml:"auth-token"`
	AuthTokenKey string `yaml:"auth-token-key"`
}

type Database struct {
	// 数据库类型
	Type string `yaml:"type"`
	// sqlite数据库文件
	Path string `yaml:"path"`
	// 用户名
	UserName string `yaml:"username"`
	// 密码
	Password string `yaml:"password"`
	// 主机
	Host string `yaml:"host"`
	// 数据库名
	Name string `yaml:"name"`
	// 表前缀
	TablePrefix string `yaml:"table-prefix"`

	// 是否启用自动迁移
	AutoMigrate bool `yaml:"auto-migrate"` // 新增字段

	// 字符集
	Charset string `yaml:"charset"`
	// 是否解析时间
	ParseTime bool `yaml:"parse-time"`
	// 最大闲置连接数
	MaxIdleConns int `yaml:"max-idle-conns"`
	// 最大打开连接数
	MaxOpenConns int `yaml:"max-open-conns"`
}

type user struct {
	// 是否启用
	IsEnabled bool `yaml:"is-enable"`
	// 注册是否启用
	RegisterIsEnable bool `yaml:"register-is-enable"`
}

type app struct {
	// 默认页面大小
	DefaultPageSize int `yaml:"default-page-size"`
	// 最大页面大小
	MaxPageSize int `yaml:"max-page-size"`
	// 默认上下文超时时间
	DefaultContextTimeout int `yaml:"default-context-timeout"`
	// 日志保存路径
	LogSavePath string `yaml:"log-save-fileurl"`
	// 日志文件名
	LogFile string `yaml:"log-file"`

	// 上传临时路径
	TempPath string `yaml:"temp-fileurl"`
	// 上传服务器URL
	UploadUrlPre string `yaml:"upload-url-pre"`
	//上传日期路径设置
	UploadDatePath string `yaml:"upload-date-path"`
	// 上传图片最大尺寸
	UploadMaxSize int `yaml:"upload-max-size"`
	// 上传图片允许的扩展名
	UploadAllowExts []string `yaml:"upload-allow-exts"`

	ImageMaxSizeWidth  int `yaml:"image-max-size-width"`
	ImageMaxSizeHeight int `yaml:"image-max-size-height"`
	ImageQuality       int `yaml:"image-quality"`
}

type email struct {
	ErrorReportEnable bool     `yaml:"error-report-enable"`
	Host              string   `yaml:"host"`
	Port              int      `yaml:"port"`
	UserName          string   `yaml:"username"`
	Password          string   `yaml:"password"`
	IsSSL             bool     `yaml:"is-ssl"`
	From              string   `yaml:"from"`
	To                []string `yaml:"to"`
}

// ConfigLoad 初始化
func ConfigLoad(f string) (string, error) {

	realpath, err := fileurl.GetAbsPath(f, "")
	if err != nil {
		return realpath, err
	}
	c := new(config)

	c.File = f
	file, err := os.ReadFile(f)
	if err != nil {
		return realpath, errors.Wrap(err, "read config file failed")
	}

	err = yaml.Unmarshal(file, c)
	if err != nil {
		return realpath, errors.Wrap(err, "parse config file failed")
	}
	Config = c
	return realpath, nil

}
