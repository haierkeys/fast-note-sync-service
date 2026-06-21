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
	"go.uber.org/zap"
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

	// fetchHTTPTimeout bounds each release-list HTTP request so a hung source
	// can't stall the whole task; the task-layer fallback then retries the other source.
	// fetchHTTPTimeout 限制单次 release 列表请求的耗时，防止某源卡死拖住整个任务；
	// task 层 fallback 会回退到另一源重试。
	fetchHTTPTimeout = 8 * time.Second
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

// releaseSource identifies which source a release list came from.
// releaseSource 标识 release 列表来自哪个源。
type releaseSource string

const (
	sourceGitHub releaseSource = "github"
	sourceCNB    releaseSource = "cnb"
)

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
	// Primary source is decided by SourceSelector (auto probes real release
	// endpoints, with latency-aware switching). The task-layer fallback below
	// guarantees that even if the probe picked the wrong source, the other one
	// is retried once before giving up.
	// 主源由 SourceSelector 决定（auto 模式探测真实 release 接口 + 延迟感知切换）。
	// 下方 task 层 fallback 保证即使探测选错源，也会回退另一源重试一次。
	githubFirst := t.app.IsPullFromGitHub()

	// Service releases (with fallback)
	// 服务端版本（带回退）
	serviceUsedGitHub, serviceReleases, serviceLink, serviceChangelog, serviceChangelogContent, err :=
		t.fetchReleasesWithFallback(ctx, GitHubServiceReleaseURL, CNBServiceReleaseURL, CNBServiceToken, true, githubFirst)
	if err != nil {
		return fmt.Errorf("fetch service releases: %w", err)
	}

	// Plugin releases (prefer the same source the service fetch actually used,
	// so service & plugin versions come from the same mirror)
	// 插件版本（优先用服务端实际命中的源，保证服务端与插件版本来自同一镜像）
	pluginReleases, pluginLink, pluginChangelog, pluginChangelogContent, perr :=
		t.fetchReleasesWithFallbackLinksOnly(ctx, GitHubPluginReleaseURL, CNBPluginReleaseURL, CNBPluginToken, false, serviceUsedGitHub)
	if perr != nil {
		// Plugin version is best-effort; don't fail the whole task if only plugin fetch errors.
		// 插件版本尽力而为：仅插件抓取失败不应导致整个任务失败。
		t.app.Logger().Warn("check_version: plugin releases fetch failed, keeping service result only", zap.Error(perr))
		pluginReleases = nil
	}

	var serviceLatest, pluginLatest string
	if len(serviceReleases) > 0 {
		serviceLatest = serviceReleases[0].Version
	}
	if len(pluginReleases) > 0 {
		pluginLatest = pluginReleases[0].Version
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
		// GithubAvailable now means "the source the version info was ACTUALLY
		// fetched from is GitHub". This keeps the upgrade download URL (decided
		// from this field in handler_admin_control.Upgrade) consistent with the
		// source that produced the version data, fixing the old probe/upgrade
		// mismatch where the probe said GitHub but the download went 404.
		// GithubAvailable 现在表示「版本信息实际命中的源是 GitHub」，使升级下载地址
		// （在 handler_admin_control.Upgrade 中由本字段决定）与产生版本数据的源一致，
		// 修复过去「探测说 GitHub 但下载 404」的不一致。
		GithubAvailable:                  serviceUsedGitHub,
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

// fetchReleasesWithFallback tries the primary source first; on failure/empty it
// falls back to the other source once. Returns the source actually used and the
// assembled release list + link/changelog for the latest release.
// fetchReleasesWithFallback 先试主源；失败或为空时回退另一源重试一次。
// 返回实际命中的源、已过滤的 release 列表，以及最新版的 link/changelog。
func (t *CheckVersionTask) fetchReleasesWithFallback(
	ctx context.Context, ghURL, cnbURL, cnbToken string, isService, githubFirst bool,
) (usedGitHub bool, releases []pkgapp.HistoricalVersion, link, changelogLink, changelogContent string, err error) {
	releases, link, changelogLink, changelogContent, usedGitHub, err = t.tryFetch(ctx, ghURL, cnbURL, cnbToken, isService, githubFirst)
	if err == nil && len(releases) > 0 {
		return usedGitHub, releases, link, changelogLink, changelogContent, nil
	}
	if err != nil {
		t.app.Logger().Warn("check_version: primary source failed, falling back",
			zap.Bool("githubFirst", githubFirst),
			zap.Bool("isService", isService),
			zap.Error(err))
	} else {
		t.app.Logger().Warn("check_version: primary source returned no releases, falling back",
			zap.Bool("githubFirst", githubFirst),
			zap.Bool("isService", isService))
	}

	// fallback to the other source
	// 回退另一源
	releases, link, changelogLink, changelogContent, usedGitHub, ferr := t.tryFetch(ctx, ghURL, cnbURL, cnbToken, isService, !githubFirst)
	if ferr != nil {
		if err == nil {
			err = ferr
		}
		return githubFirst, nil, "", "", "", fmt.Errorf("both sources failed (primary err fallback): %w", ferr)
	}
	return usedGitHub, releases, link, changelogLink, changelogContent, nil
}

// fetchReleasesWithFallbackOnly is the variant returning links-only without the
// usedGitHub being needed by callers that ignore it. Kept the signature explicit
// for clarity in Run(); see fetchReleasesWithFallback for behavior.
// fetchReleasesWithFallbackLinksOnly 与 fetchReleasesWithFallback 行为相同，
// 供不需要 usedGitHub 的调用点使用（语义清晰）。
func (t *CheckVersionTask) fetchReleasesWithFallbackLinksOnly(
	ctx context.Context, ghURL, cnbURL, cnbToken string, isService, githubFirst bool,
) ([]pkgapp.HistoricalVersion, string, string, string, error) {
	_, releases, link, cl, clc, err := t.fetchReleasesWithFallback(ctx, ghURL, cnbURL, cnbToken, isService, githubFirst)
	return releases, link, cl, clc, err
}

// tryFetch fetches from exactly one source (github or cnb) and assembles links.
// tryFetch 只从一个源抓取（github 或 cnb）并拼装链接。
func (t *CheckVersionTask) tryFetch(
	ctx context.Context, ghURL, cnbURL, cnbToken string, isService, useGitHub bool,
) (releases []pkgapp.HistoricalVersion, link, changelogLink, changelogContent string, usedGitHub bool, err error) {
	if useGitHub {
		releases, err = t.fetchGitHubReleasesCtx(ctx, ghURL)
	} else {
		releases, err = t.fetchCNBVersionCtx(ctx, cnbURL, cnbToken)
	}
	if err != nil {
		return nil, "", "", "", useGitHub, err
	}
	link, changelogLink, changelogContent = buildLinks(releases, isService, useGitHub)
	return releases, link, changelogLink, changelogContent, useGitHub, nil
}

// buildLinks assembles the release page / changelog URLs for the latest release,
// depending on which source it came from.
// buildLinks 根据来源源，为最新版拼装 release 页面与 changelog 的 URL。
func buildLinks(releases []pkgapp.HistoricalVersion, isService, fromGitHub bool) (link, changelogLink, changelogContent string) {
	if len(releases) == 0 {
		return "", "", ""
	}
	latest := releases[0].Version
	changelogContent = releases[0].ChangelogContent
	latestClean := strings.TrimPrefix(latest, "v")
	if fromGitHub {
		base := ServiceRepoURL
		if !isService {
			base = PluginRepoURL
		}
		link = base + "/releases/tag/" + latestClean
		changelogLink = base + "/releases/download/" + latestClean + "/changelog.txt"
	} else {
		base := CNBServiceURL
		if !isService {
			base = CNBPluginURL
		}
		link = base + "/-/releases/tag/" + latestClean
		changelogLink = base + "/-/releases/download/" + latestClean + "/changelog.txt"
	}
	return link, changelogLink, changelogContent
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

func (t *CheckVersionTask) fetchGitHubReleasesCtx(ctx context.Context, url string) ([]pkgapp.HistoricalVersion, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	cli := &http.Client{Timeout: fetchHTTPTimeout}
	resp, err := cli.Do(req)
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

func (t *CheckVersionTask) fetchCNBVersionCtx(ctx context.Context, url string, token string) ([]pkgapp.HistoricalVersion, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.cnb.api+json")
	req.Header.Set("Authorization", token)

	cli := &http.Client{Timeout: fetchHTTPTimeout}
	resp, err := cli.Do(req)
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
