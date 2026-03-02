package fileurl

import (
	"net/http"
	"time"
)

const (
	SourceAuto   = "auto"
	SourceGitHub = "github"
	SourceCNB    = "cnb"

	GitHubCheckURL = "https://api.github.com"
)

// SourceSelector handles the logic of selecting the data source (GitHub or CNB)
// SourceSelector 处理选择数据源（GitHub 或 CNB）的逻辑
type SourceSelector struct {
	mode      string
	isGitHub  bool
	lastCheck time.Time
}

// NewSourceSelector creates a new SourceSelector
// NewSourceSelector 创建一个新的 SourceSelector
func NewSourceSelector(mode string) *SourceSelector {
	return &SourceSelector{
		mode: mode,
	}
}

// IsGitHub returns whether the current source should be GitHub
// IsGitHub 返回当前源是否应该为 GitHub
func (s *SourceSelector) IsGitHub() bool {
	switch s.mode {
	case SourceGitHub:
		return true
	case SourceCNB:
		return false
	default:
		// Auto mode or unknown mode
		// 自动模式或未知模式
		if time.Since(s.lastCheck) > 10*time.Minute {
			s.isGitHub = s.checkGitHub()
			s.lastCheck = time.Now()
		}
		return s.isGitHub
	}
}

func (s *SourceSelector) checkGitHub() bool {
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Head(GitHubCheckURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
