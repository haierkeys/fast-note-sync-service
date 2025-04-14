// webdav.go

package webdav

import (
	"github.com/studio-b12/gowebdav"
)

// Config 结构体用于存储 WebDAV 连接信息。
type Config struct {
	IsEnabled  bool   `yaml:"is-enable"`
	Endpoint   string `yaml:"endpoint"`
	Path       string `yaml:"path"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	CustomPath string `yaml:"custom-path"`
}

// WebDAV 结构体表示 WebDAV 客户端。
type WebDAV struct {
	Client *gowebdav.Client
	Config *Config
}

var clients = make(map[string]*WebDAV)

// NewClient 创建一个新的 WebDAV 客户端实例。
func NewClient(cf map[string]any) (*WebDAV, error) {
	// New client

	var IsEnabled bool
	switch t := cf["IsEnabled"].(type) {
	case int64:
		if t == 0 {
			IsEnabled = false
		} else {
			IsEnabled = true
		}
	case bool:
		IsEnabled = t
	}

	conf := &Config{
		IsEnabled:  IsEnabled,
		Endpoint:   cf["Endpoint"].(string),
		User:       cf["User"].(string),
		Password:   cf["Password"].(string),
		CustomPath: cf["CustomPath"].(string),
	}

	var endpoint = conf.Endpoint
	var path = conf.Path
	var user = conf.User
	var password = conf.Password
	var customPath = conf.CustomPath

	if clients[endpoint+path+user+customPath] != nil {
		return clients[endpoint+path+user+customPath], nil
	}

	c := gowebdav.NewClient(endpoint, user, password)
	c.Connect()

	clients[endpoint+path+user+customPath] = &WebDAV{
		Client: c,
		Config: conf,
	}
	return clients[endpoint+path+user+customPath], nil
}
