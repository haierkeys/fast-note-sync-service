package task

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"golang.org/x/mod/semver"
)

const (
	GitHubCheckURL    = "https://api.github.com"
	ServiceRepoPath   = "haierkeys/fast-note-sync-service"
	ServiceRepoURL    = "https://github.com/" + ServiceRepoPath
	PluginRepoPath    = "haierkeys/obsidian-fast-note-sync"
	PluginRepoURL     = "https://github.com/" + PluginRepoPath
	ServiceVersionURL = "https://img.shields.io/github/v/release/" + ServiceRepoPath + ".json"
	PluginVersionURL  = "https://img.shields.io/github/v/tag/" + PluginRepoPath + ".json"

	CNBServiceReleaseURL = "https://api.cnb.cool/" + ServiceRepoPath + "/-/releases"
	CNBPluginReleaseURL  = "https://api.cnb.cool/" + PluginRepoPath + "/-/releases"
	CNBServiceURL        = "https://cnb.cool/" + ServiceRepoPath
	CNBPluginURL         = "https://cnb.cool/" + PluginRepoPath
)

type CNBRelease struct {
	TagName string `json:"tag_name"`
}

type ShieldsJSON struct {
	Message string `json:"message"`
}

type CheckVersionTask struct {
	app      *app.App
	isGitHub bool
}

func init() {
	RegisterWithApp(func(appContainer *app.App) (Task, error) {
		return &CheckVersionTask{
			app: appContainer,
		}, nil
	})
}

func (t *CheckVersionTask) Name() string {
	return "check_version"
}

func (t *CheckVersionTask) Run(ctx context.Context) error {
	t.isGitHub = t.checkGitHub()

	var serviceLatest, pluginLatest string
	var serviceLink, pluginLink string
	var serviceChangelog, pluginChangelog string
	var err error

	if t.isGitHub {
		serviceLatest, err = t.fetchVersion(ServiceVersionURL)
		if err != nil {
			return err
		}
		pluginLatest, err = t.fetchVersion(PluginVersionURL)
		if err != nil {
			return err
		}
		serviceLink = ServiceRepoURL + "/releases/tag/" + serviceLatest
		pluginLink = PluginRepoURL + "/releases/tag/" + pluginLatest
		serviceChangelog = ServiceRepoURL + "/releases/download/" + serviceLatest + "/changelog.txt"
		pluginChangelog = PluginRepoURL + "/releases/download/" + pluginLatest + "/changelog.txt"
	} else {
		serviceLatest, err = t.fetchCNBVersion(CNBServiceReleaseURL)
		if err != nil {
			return err
		}
		pluginLatest, err = t.fetchCNBVersion(CNBPluginReleaseURL)
		if err != nil {
			return err
		}
		serviceLink = CNBServiceURL + "/-/releases/tag/" + serviceLatest
		pluginLink = CNBPluginURL + "/-/releases/tag/" + pluginLatest
		serviceChangelog = CNBServiceURL + "/-/releases/download/" + serviceLatest + "/changelog.txt"
		pluginChangelog = CNBPluginURL + "/-/releases/download/" + pluginLatest + "/changelog.txt"
	}

	currentServiceVersion := t.app.Version().Version
	if !strings.HasPrefix(currentServiceVersion, "v") {
		currentServiceVersion = "v" + currentServiceVersion
	}

	if !strings.HasPrefix(serviceLatest, "v") {
		serviceLatest = "v" + serviceLatest
	}

	if !strings.HasPrefix(pluginLatest, "v") {
		pluginLatest = "v" + pluginLatest
	}

	info := pkgapp.CheckVersionInfo{
		GithubAvailable:           t.isGitHub,
		VersionNewName:            serviceLatest,
		VersionIsNew:              semver.Compare(serviceLatest, currentServiceVersion) > 0,
		VersionNewLink:            serviceLink,
		VersionNewChangelog:       serviceChangelog,
		PluginVersionNewName:      pluginLatest,
		PluginVersionNewLink:      pluginLink,
		PluginVersionNewChangelog: pluginChangelog,
		// 这里无法判断 PluginVersionIsNew，因为没有具体的客户端版本，
		// 但我们还是更新最新的版本名，具体的比较逻辑可以在 App.CheckVersion 中根据传入的参数进行。
	}

	// 更新 App 中的版本信息
	t.app.SetCheckVersionInfo(info)

	return nil
}

func (t *CheckVersionTask) fetchVersion(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var sj ShieldsJSON
	if err := json.Unmarshal(body, &sj); err != nil {
		return "", err
	}

	return sj.Message, nil
}

func (t *CheckVersionTask) fetchCNBVersion(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var releases []CNBRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", err
	}

	if len(releases) == 0 {
		return "", nil
	}

	return releases[0].TagName, nil
}

func (t *CheckVersionTask) checkGitHub() bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Head(GitHubCheckURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (t *CheckVersionTask) LoopInterval() time.Duration {
	return 30 * time.Minute
}

func (t *CheckVersionTask) IsStartupRun() bool {
	return true
}
