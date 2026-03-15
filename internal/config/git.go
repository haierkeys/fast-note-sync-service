package config

// GitConfig Git commit configuration
// GitConfig Git 提交配置
type GitConfig struct {
	// Name author name for git commit
	// Name git 提交的作者名称
	Name string `yaml:"name" default:"FNS Service"`
	// Email author email for git commit
	// Email git 提交的作者邮箱
	Email string `yaml:"email" default:"fns@email.com"`
}
