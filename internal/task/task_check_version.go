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

type GitHubAsset struct {
	Name  string `json:"name"`  // Asset name // 资源包名称
	State string `json:"state"` // Upload state // 上传状态
}

type CNBRelease struct {
	TagName    string        `json:"tag_name"`
	Prerelease bool          `json:"prerelease"`
	Body       string        `json:"body"`   // Release description (changelog) // 版本发布说明（更新日志）
	Assets     []GitHubAsset `json:"assets"` // Release assets // 资源列表
}

type GitHubRelease struct {
	TagName    string        `json:"tag_name"`
	Prerelease bool          `json:"prerelease"`
	Body       string        `json:"body"`   // Release description (changelog) // 版本发布说明（更新日志）
	Assets     []GitHubAsset `json:"assets"` // Release assets // 资源列表
}

type GitHubTag struct {
	Name string `json:"name"`
}

type CheckVersionTask struct {
	app *app.App
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
	isGitHub := t.app.IsPullFromGitHub()

	var serviceLatest, pluginLatest string
	var serviceLink, pluginLink string
	var serviceChangelog, pluginChangelog string
	var err error

	var serviceChangelogContent, pluginChangelogContent string
	var serviceReleases, pluginReleases []pkgapp.HistoricalVersion

	if isGitHub {
		serviceReleases, err = t.fetchGitHubReleases(GitHubServiceReleaseURL)
		if err != nil {
			return err
		}

		pluginReleases, err = t.fetchGitHubReleases(GitHubPluginReleaseURL)
		if err != nil {
			return err
		}

		if len(serviceReleases) > 0 {
			serviceLatest = serviceReleases[0].Version
			serviceChangelogContent = serviceReleases[0].ChangelogContent
			serviceLatestClean := strings.TrimPrefix(serviceLatest, "v")
			serviceLink = ServiceRepoURL + "/releases/tag/" + serviceLatestClean
			serviceChangelog = ServiceRepoURL + "/releases/download/" + serviceLatestClean + "/changelog.txt"
		}

		if len(pluginReleases) > 0 {
			pluginLatest = pluginReleases[0].Version
			pluginChangelogContent = pluginReleases[0].ChangelogContent
			pluginLatestClean := strings.TrimPrefix(pluginLatest, "v")
			pluginLink = PluginRepoURL + "/releases/tag/" + pluginLatestClean
			pluginChangelog = PluginRepoURL + "/releases/download/" + pluginLatestClean + "/changelog.txt"
		}

	} else {
		serviceReleases, err = t.fetchCNBVersion(CNBServiceReleaseURL, CNBServiceToken)
		if err != nil {
			return err
		}
		pluginReleases, err = t.fetchCNBVersion(CNBPluginReleaseURL, CNBPluginToken)
		if err != nil {
			return err
		}

		if len(serviceReleases) > 0 {
			serviceLatest = serviceReleases[0].Version
			serviceChangelogContent = serviceReleases[0].ChangelogContent
			serviceLatestClean := strings.TrimPrefix(serviceLatest, "v")
			serviceLink = CNBServiceURL + "/-/releases/tag/" + serviceLatestClean
			serviceChangelog = CNBServiceURL + "/-/releases/download/" + serviceLatestClean + "/changelog.txt"
		}

		if len(pluginReleases) > 0 {
			pluginLatest = pluginReleases[0].Version
			pluginChangelogContent = pluginReleases[0].ChangelogContent
			pluginLatestClean := strings.TrimPrefix(pluginLatest, "v")
			pluginLink = CNBPluginURL + "/-/releases/tag/" + pluginLatestClean
			pluginChangelog = CNBPluginURL + "/-/releases/download/" + pluginLatestClean + "/changelog.txt"
		}
	}

	currentServiceVersion := t.app.Version().Version
	if !strings.HasPrefix(currentServiceVersion, "v") {
		currentServiceVersion = "v" + currentServiceVersion
	}

	if serviceLatest != "" && !strings.HasPrefix(serviceLatest, "v") {
		serviceLatest = "v" + serviceLatest
	}

	if pluginLatest != "" && !strings.HasPrefix(pluginLatest, "v") {
		pluginLatest = "v" + pluginLatest
	}

	info := pkgapp.CheckVersionInfo{
		GithubAvailable:                  isGitHub,
		VersionNewName:                   serviceLatest,
		VersionIsNew:                     serviceLatest != "" && semver.Compare(serviceLatest, currentServiceVersion) > 0,
		VersionNewLink:                   serviceLink,
		VersionNewChangelog:              serviceChangelog,
		VersionNewChangelogContent:       serviceChangelogContent,
		PluginVersionNewName:             pluginLatest,
		PluginVersionNewLink:             pluginLink,
		PluginVersionNewChangelog:        pluginChangelog,
		PluginVersionNewChangelogContent: pluginChangelogContent,
	}

	// 更新 App 中的版本信息和发布列表
	t.app.SetCheckVersionInfo(info)
	t.app.SetCheckVersionReleases(serviceReleases, pluginReleases)

	// 推送版本信息给所有已连接客户端
	t.app.BroadcastClientInfo()

	return nil
}

// hasValidAssets checks if there is at least one uploaded zip or tar.gz file
// hasValidAssets 检查是否包含至少一个已成功上传的 zip 或 tar.gz 资源文件
func hasValidAssets(assets []GitHubAsset) bool {
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".tar.gz") {
			if asset.State == "" || asset.State == "uploaded" {
				return true
			}
		}
	}
	return false
}

func (t *CheckVersionTask) fetchGitHubReleases(url string) ([]pkgapp.HistoricalVersion, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}

	releaseChannel := t.app.Config().App.PullReleaseChannel
	var result []pkgapp.HistoricalVersion
	for _, release := range releases {
		if releaseChannel == "stable" && release.Prerelease {
			continue
		}
		if !hasValidAssets(release.Assets) {
			continue
		}
		tagName := release.TagName
		if !strings.HasPrefix(tagName, "v") {
			tagName = "v" + tagName
		}
		result = append(result, pkgapp.HistoricalVersion{
			Version:          tagName,
			ChangelogContent: release.Body,
		})
	}

	return result, nil
}

func (t *CheckVersionTask) fetchCNBVersion(url string, token string) ([]pkgapp.HistoricalVersion, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.cnb.api+json")
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var releases []CNBRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}

	releaseChannel := t.app.Config().App.PullReleaseChannel
	var result []pkgapp.HistoricalVersion
	for _, release := range releases {
		// CNB API usually follows Gitea/GitHub pattern
		// Also fallback check for common prerelease suffixes if field is not enough
		isPrerelease := release.Prerelease
		if !isPrerelease {
			tagName := strings.ToLower(release.TagName)
			if strings.Contains(tagName, "-beta") || strings.Contains(tagName, "-rc") || strings.Contains(tagName, "-alpha") {
				isPrerelease = true
			}
		}

		if releaseChannel == "stable" && isPrerelease {
			continue
		}
		if !hasValidAssets(release.Assets) {
			continue
		}
		tagName := release.TagName
		if !strings.HasPrefix(tagName, "v") {
			tagName = "v" + tagName
		}
		result = append(result, pkgapp.HistoricalVersion{
			Version:          tagName,
			ChangelogContent: release.Body,
		})
	}

	return result, nil
}

func (t *CheckVersionTask) LoopInterval() time.Duration {
	return 10 * time.Minute
}

func (t *CheckVersionTask) IsStartupRun() bool {
	return true
}
