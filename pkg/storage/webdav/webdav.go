package webdav

import (
	"github.com/studio-b12/gowebdav"
)

// Config 结构体用于存储 WebDAV 连接信息。
type Config struct {
	IsEnabled     bool   `yaml:"is-enable"`
	IsUserEnabled bool   `yaml:"is-user-enable"`
	Endpoint      string `yaml:"endpoint"`
	Path          string `yaml:"path"`
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	CustomPath    string `yaml:"custom-path"`
}

// WebDAV 结构体表示 WebDAV 客户端。
type WebDAV struct {
	Client *gowebdav.Client
	Config *Config
}

var clients = make(map[string]*WebDAV)

// NewClient 创建一个新的 WebDAV 客户端实例。
func NewClient(conf *Config) (*WebDAV, error) {
	var endpoint = conf.Endpoint
	var path = conf.Path
	var user = conf.User
	var customPath = conf.CustomPath

	if clients[endpoint+path+user+customPath] != nil {
		return clients[endpoint+path+user+customPath], nil
	}

	c := gowebdav.NewClient(endpoint, user, conf.Password)
	c.Connect()

	clients[endpoint+path+user+customPath] = &WebDAV{
		Client: c,
		Config: conf,
	}
	return clients[endpoint+path+user+customPath], nil
}
