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

// Support retrieves support records by language with pagination and sorting
// Support 分页并排序获取指定语言的打赏记录
// @Summary Get support records
// @Description Get support records for the specified language with pagination and sorting
// @Tags System
// @Produce json
// @Param lang query string false "Language code (default: en)"
// @Param sortBy query string false "Sort by field (amount, time, name, item)"
// @Param sortOrder query string false "Sort order (asc, desc)"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {object} pkgapp.Res{data=pkgapp.ListRes} "Success"
// @Router /api/support [get]
func (h *VersionHandler) Support(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	lang := strings.ToLower(c.Query("lang"))
	if lang == "" {
		lang = "en"
	}

	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	pager := pkgapp.NewPager(c)
	data, total := h.App.GetSupportRecordsPage(lang, sortBy, sortOrder, pager.Page, pager.PageSize)

	response.ToResponseList(code.Success, data, total)
}
