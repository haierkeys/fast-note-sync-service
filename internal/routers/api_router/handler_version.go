package api_router

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
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
		Version:             versionInfo.Version,
		GitTag:              versionInfo.GitTag,
		BuildTime:           versionInfo.BuildTime,
		VersionIsNew:        checkInfo.VersionIsNew,
		VersionNewName:      checkInfo.VersionNewName,
		VersionNewLink:      checkInfo.VersionNewLink,
		VersionNewChangelog: checkInfo.VersionNewChangelog,
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
