package task

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"golang.org/x/mod/semver"
)

const (
	GitHubCheckURL          = "https://api.github.com"
	GitHubServiceReleaseURL = "https://api.github.com/repos/haierkeys/fast-note-sync-service/releases"
	GitHubPluginReleaseURL  = "https://api.github.com/repos/haierkeys/obsidian-fast-note-sync/releases"
	ServiceRepoPath         = "haierkeys/fast-note-sync-service"
	ServiceRepoURL          = "https://github.com/" + ServiceRepoPath
	PluginRepoPath          = "haierkeys/obsidian-fast-note-sync"
	PluginRepoURL           = "https://github.com/" + PluginRepoPath

	CNBServiceReleaseURL = "https://api.cnb.cool/" + ServiceRepoPath + "/-/releases"
	CNBPluginReleaseURL  = "https://api.cnb.cool/" + PluginRepoPath + "/-/releases"
	CNBServiceURL        = "https://cnb.cool/" + ServiceRepoPath
	CNBPluginURL         = "https://cnb.cool/" + PluginRepoPath
	CNBServiceToken      = "58tjez3744HL9Z10cRaCHdeEPhK"
	CNBPluginToken       = "9pFNKcjlej36e0w6MHKT6YMn53G"
)

type CNBRelease struct {
	TagName string `json:"tag_name"`
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

type GitHubTag struct {
	Name string `json:"name"`
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
		serviceLatest, err = t.fetchGitHubReleases(GitHubServiceReleaseURL)
		if err != nil {
			return err
		}

		pluginLatest, err = t.fetchGitHubReleases(GitHubPluginReleaseURL)
		if err != nil {
			return err
		}
		serviceLink = ServiceRepoURL + "/releases/tag/" + serviceLatest
		pluginLink = PluginRepoURL + "/releases/tag/" + pluginLatest
		serviceChangelog = ServiceRepoURL + "/releases/download/" + serviceLatest + "/changelog.txt"
		pluginChangelog = PluginRepoURL + "/releases/download/" + pluginLatest + "/changelog.txt"

	} else {
		serviceLatest, err = t.fetchCNBVersion(CNBServiceReleaseURL, CNBServiceToken)
		if err != nil {
			return err
		}
		pluginLatest, err = t.fetchCNBVersion(CNBPluginReleaseURL, CNBPluginToken)
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
		GithubAvailable:                  t.isGitHub,
		VersionNewName:                   serviceLatest,
		VersionIsNew:                     semver.Compare(serviceLatest, currentServiceVersion) > 0,
		VersionNewLink:                   serviceLink,
		VersionNewChangelog:              serviceChangelog,
		VersionNewChangelogContent:       t.fetchTextContent(serviceChangelog),
		PluginVersionNewName:             pluginLatest,
		PluginVersionNewLink:             pluginLink,
		PluginVersionNewChangelog:        pluginChangelog,
		PluginVersionNewChangelogContent: t.fetchTextContent(pluginChangelog),
	}

	// 更新 App 中的版本信息
	t.app.SetCheckVersionInfo(info)

	return nil
}

func (t *CheckVersionTask) fetchGitHubReleases(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", err
	}

	if len(releases) == 0 {
		return "", nil
	}

	return strings.TrimPrefix(releases[0].TagName, "v"), nil
}

func (t *CheckVersionTask) fetchCNBVersion(url string, token string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.cnb.api+json")
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
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

	return strings.TrimPrefix(releases[0].TagName, "v"), nil
}

func (t *CheckVersionTask) fetchTextContent(url string) string {
	if url == "" {
		return ""
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(body)
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
