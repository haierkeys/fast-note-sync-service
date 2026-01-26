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
	ServiceVersionURL = "https://img.shields.io/github/v/release/haierkeys/fast-note-sync-service.json"
	PluginVersionURL  = "https://img.shields.io/github/v/tag/haierkeys/obsidian-fast-note-sync.json"
)

type ShieldsJSON struct {
	Message string `json:"message"`
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
	serviceLatest, err := t.fetchVersion(ServiceVersionURL)
	if err != nil {
		return err
	}

	pluginLatest, err := t.fetchVersion(PluginVersionURL)
	if err != nil {
		return err
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
		VersionNewName:       serviceLatest,
		VersionIsNew:         semver.Compare(serviceLatest, currentServiceVersion) > 0,
		PluginVersionNewName: pluginLatest,
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

func (t *CheckVersionTask) LoopInterval() time.Duration {
	return 30 * time.Minute
}

func (t *CheckVersionTask) IsStartupRun() bool {
	return true
}
