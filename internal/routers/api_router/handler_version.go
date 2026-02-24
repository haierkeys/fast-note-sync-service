package api_router

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

// VersionHandler version info API router handler
// VersionHandler 版本信息 API 路由处理器
// Uses App Container to inject dependencies
// 使用 App Container 注入依赖
type VersionHandler struct {
	*Handler
}

// NewVersionHandler creates VersionHandler instance
// NewVersionHandler 创建 VersionHandler 实例
func NewVersionHandler(a *app.App) *VersionHandler {
	return &VersionHandler{
		Handler: NewHandler(a),
	}
}

// ServerVersion retrieves server version information
// @Summary Get server version info
// @Description Get current server software version, Git tag, and build time
// @Tags System
// @Produce json
// @Success 200 {object} pkgapp.Res{data=dto.VersionDTO} "Success"
// @Router /api/version [get]
func (h *VersionHandler) ServerVersion(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	versionInfo := h.App.Version()
	checkInfo := h.App.CheckVersion("")
	response.ToResponse(code.Success.WithData(dto.VersionDTO{
		Version:                          versionInfo.Version,
		GitTag:                           versionInfo.GitTag,
		BuildTime:                        versionInfo.BuildTime,
		VersionIsNew:                     checkInfo.VersionIsNew,
		VersionNewName:                   checkInfo.VersionNewName,
		VersionNewLink:                   checkInfo.VersionNewLink,
		VersionNewChangelog:              checkInfo.VersionNewChangelog,
		VersionNewChangelogContent:       checkInfo.VersionNewChangelogContent,
		PluginVersionNewName:             checkInfo.PluginVersionNewName,
		PluginVersionNewLink:             checkInfo.PluginVersionNewLink,
		PluginVersionNewChangelog:        checkInfo.PluginVersionNewChangelog,
		PluginVersionNewChangelogContent: checkInfo.PluginVersionNewChangelogContent,
	}))
}

// Support retrieves support records by language
// @Summary Get support records
// @Description Get support records for the specified language
// @Tags System
// @Produce json
// @Param lang query string false "Language code (default: en)"
// @Success 200 {object} pkgapp.Res{data=[]pkgapp.SupportRecord} "Success"
// @Router /api/support [get]
func (h *VersionHandler) Support(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	lang := strings.ToLower(c.Query("lang"))
	if lang == "" {
		lang = "en"
	}

	records := h.App.GetSupportRecords()
	data, ok := records[lang]
	if !ok {
		// Fallback to en if requested language is not found
		// 如果找不到请求的语言，回退到 en
		data = records["en"]
	}

	if data == nil {
		data = []pkgapp.SupportRecord{}
	}

	response.ToResponse(code.Success.WithData(data))
}

// Upgrade triggers server automatic upgrade
// @Summary Trigger server upgrade
// @Description Download latest version and restart server
// @Tags System
// @Produce json
// @Security UserAuthToken
// @Param version query string true "Version to upgrade (e.g. 2.0.10 or latest)"
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/admin/upgrade [get]
func (h *VersionHandler) Upgrade(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	uid := pkgapp.GetUID(c)

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	var upgradeReq dto.UpgradeRequest
	if ok, validErrs := pkgapp.BindAndValid(c, &upgradeReq); !ok {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(validErrs.Errors()...))
		return
	}

	checkInfo := h.App.CheckVersion("")
	version := ""

	if upgradeReq.Version == "latest" {
		if !checkInfo.VersionIsNew {
			response.ToResponse(code.Success.WithDetails("Current version is already up to date"))
			return
		}
		version = checkInfo.VersionNewName
	} else {
		version = upgradeReq.Version
	}

	versionRaw := strings.TrimPrefix(version, "v")

	// Determine download URL
	// 确定下载地址
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Example: fast-note-sync-service-2.0.10-linux-amd64.tar.gz
	fileName := fmt.Sprintf("fast-note-sync-service-%s-%s-%s.tar.gz", versionRaw, goos, goarch)
	downloadURL := ""
	if checkInfo.GithubAvailable {
		// GitHub releases/download/[tag]/[filename]
		// Based on user feedback: URL should NOT have 'v' in the tag part if the tag itself doesn't have it
		downloadURL = fmt.Sprintf("https://github.com/haierkeys/fast-note-sync-service/releases/download/%s/%s", versionRaw, fileName)
	} else {
		// CNB download URL format
		downloadURL = fmt.Sprintf("https://cnb.cool/haierkeys/fast-note-sync-service/-/releases/download/%s/%s", versionRaw, fileName)
	}

	h.App.Logger().Info("Starting upgrade download", zap.String("url", downloadURL), zap.String("version", versionRaw))

	// Prepare temp directory
	// 使用 storage/temp/upgrade 作为临时目录
	tempDir := filepath.Join("storage", "temp", "upgrade")
	_ = os.RemoveAll(tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		response.ToResponse(code.Failed.WithDetails("Failed to create temp directory: " + err.Error()))
		return
	}

	// Download
	tarPath := filepath.Join(tempDir, fileName)
	if err := h.downloadFile(downloadURL, tarPath); err != nil {
		response.ToResponse(code.Failed.WithDetails("Download failed: " + err.Error()))
		return
	}

	// Extract
	binaryName := "fast-note-sync-service"
	if goos == "windows" {
		binaryName += ".exe"
	}
	extractedBinaryPath := filepath.Join(tempDir, binaryName)

	if err := h.extractBinary(tarPath, tempDir, binaryName); err != nil {
		response.ToResponse(code.Failed.WithDetails("Extract failed: " + err.Error()))
		return
	}

	// Trigger upgrade in App
	h.App.TriggerUpgrade(extractedBinaryPath)

	response.ToResponse(code.Success.WithDetails("Upgrade triggered, server is restarting..."))
}

func (h *VersionHandler) downloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (h *VersionHandler) extractBinary(tarPath string, destDir string, binaryName string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Check if it's the binary we're looking for
		// Often files in tar.gz are in a subdirectory or have different names
		// In alpha-release.yml: tar -czvf ... . (contents of build/platform dir)
		if filepath.Base(header.Name) == binaryName {
			target := filepath.Join(destDir, binaryName)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
			return nil
		}
	}

	return fmt.Errorf("binary %s not found in archive", binaryName)
}
