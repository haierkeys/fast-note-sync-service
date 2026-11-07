package mcp_router

import (
	"context"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

// Page-size bounds shared by the query tools.
// 查询工具共用的分页边界。
const (
	defaultQueryLimit = 20
	maxQueryLimit     = 100
)

// registerSearchTools wires the knowledge-query tools: search, link-graph
// integrity, tags, recency, folder listing and vault statistics.
//
// These exist because note_list can only do path/substring matching and always
// returns whole note bodies. An agent asking "where did I write about X" would
// either get nothing back or fill its context window with one note. Every tool
// below answers a structural question the note CRUD surface cannot express, and
// returns metadata plus a snippet rather than bodies — call note_get for a full
// read once you know the path.
//
// registerSearchTools 注册知识库查询工具：搜索、链接图完整性、标签、最近更新、
// 目录清单与库统计。
// 存在理由：note_list 只能做路径/子串匹配且总是返回整篇正文。Agent 问
// “我在哪写过 X”时要么搜不到，要么被一篇笔记撑满上下文。
// 下列工具覆盖笔记 CRUD 接口无法表达的结构性问题，且只返回元数据与片段而非正文 ——
// 确认路径后再用 note_get 读全文。
func registerSearchTools(srv *mcpsrv.MCPServer, appContainer *app.App) {
	cfg := appContainer.Config()

	// 1. note_search -------------------------------------------------------
	toolSearch := mcp.NewTool("note_search",
		mcp.WithDescription("Full-text search across a vault. Returns note metadata plus a short snippet around the first match instead of whole bodies, and the true match total for pagination. Use this before note_get to find notes by content."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("keyword", mcp.Required(), mcp.Description("Keyword or phrase to search for")),
		mcp.WithString("mode", mcp.Description("'fulltext' (default) searches note bodies through the full-text index; 'path' matches note paths only.")),
		mcp.WithNumber("page", mcp.Description("Page number, 1-based. Default 1.")),
		mcp.WithNumber("pageSize", mcp.Description("Hits per page, default 20, max 100.")),
		mcp.WithOutputSchema[dto.McpNoteSearchOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolSearch, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault := resolveVaultArg(ctx, appContainer, args)
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		keyword, _ := args["keyword"].(string)
		if strings.TrimSpace(keyword) == "" {
			return mcp.NewToolResultError("keyword is required"), nil
		}
		searchMode, outMode := "content", "fulltext"
		if mode, _ := args["mode"].(string); strings.EqualFold(mode, "path") {
			searchMode, outMode = "path", "path"
		}
		page, pageSize := pageArgs(args, defaultQueryLimit)

		vaultID, err := appContainer.VaultService.MustGetID(ctx, uid, vault)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		hits, total, err := appContainer.NoteRepo.SearchHits(ctx, vaultID, uid, keyword, searchMode, false, "mtime", "desc", pageSize, (page-1)*pageSize)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		out := &dto.McpNoteSearchOutput{
			Vault:    vault,
			Keyword:  keyword,
			Mode:     outMode,
			Count:    len(hits),
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			HasMore:  page*pageSize < total,
			Notes:    make([]*dto.McpNoteHitDTO, 0, len(hits)),
		}
		for _, h := range hits {
			if h.Note == nil {
				continue
			}
			out.Notes = append(out.Notes, &dto.McpNoteHitDTO{
				Path:             h.Note.Path,
				PathHash:         h.Note.PathHash,
				Snippet:          h.Snippet,
				MatchLine:        matchLineOf(h.Note.Content, keyword),
				Size:             h.Note.Size,
				Version:          h.Note.Version,
				Mtime:            h.Note.Mtime,
				UpdatedTimestamp: h.Note.UpdatedTimestamp,
				UpdatedAt:        formatNoteTime(h.Note.UpdatedAt),
			})
		}
		return mcp.NewToolResultStructured(out, searchSummary(out)), nil
	})

	// 2. note_list_broken_links -------------------------------------------
	toolBroken := mcp.NewTool("note_list_broken_links",
		mcp.WithDescription("List wikilinks whose target note does not exist in the vault, together with the note that contains each link. This is the vault integrity check: it finds links left pointing at renamed or deleted notes."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithNumber("limit", mcp.Description("Maximum broken links to return, default 100.")),
		mcp.WithOutputSchema[dto.McpNoteBrokenLinksOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolBroken, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault := resolveVaultArg(ctx, appContainer, args)
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		limit := intArg(args, "limit", 100)

		notes, links, err := loadGraph(ctx, appContainer, vault, uid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		pathByID, byHash, byBase := buildNoteIndexes(notes)

		out := &dto.McpNoteBrokenLinksOutput{
			Vault:      vault,
			TotalLinks: len(links),
			Broken:     make([]*dto.McpNoteBrokenLinkDTO, 0, 16),
		}
		for _, l := range links {
			if _, ok := resolveLinkPath(l.TargetPath, l.TargetPathHash, byHash, byBase); ok {
				continue
			}
			out.Broken = append(out.Broken, &dto.McpNoteBrokenLinkDTO{
				SourcePath: pathByID[l.SourceNoteID],
				LinkText:   l.TargetPath,
			})
			if limit > 0 && len(out.Broken) >= limit {
				break
			}
		}
		out.Count = len(out.Broken)
		return mcp.NewToolResultStructured(out, brokenLinksSummary(out)), nil
	})

	// 3. note_list_orphans -------------------------------------------------
	toolOrphans := mcp.NewTool("note_list_orphans",
		mcp.WithDescription("List notes that no other note links to. Orphans are notes no longer reachable from any entry point. Pair with note_list_broken_links to see both directions of a broken link graph."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithNumber("limit", mcp.Description("Maximum orphan notes to return, default 100.")),
		mcp.WithOutputSchema[dto.McpNoteOrphanListOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolOrphans, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault := resolveVaultArg(ctx, appContainer, args)
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		limit := intArg(args, "limit", 100)

		notes, links, err := loadGraph(ctx, appContainer, vault, uid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		_, byHash, byBase := buildNoteIndexes(notes)
		linked := make(map[string]bool, len(links))
		for _, l := range links {
			if p, ok := resolveLinkPath(l.TargetPath, l.TargetPathHash, byHash, byBase); ok {
				linked[p] = true
			}
		}

		out := &dto.McpNoteOrphanListOutput{
			Vault:      vault,
			TotalNotes: len(notes),
			Notes:      make([]string, 0, 16),
		}
		for _, n := range notes {
			if linked[n.Path] {
				continue
			}
			out.Notes = append(out.Notes, n.Path)
			if limit > 0 && len(out.Notes) >= limit {
				break
			}
		}
		out.Count = len(out.Notes)
		return mcp.NewToolResultStructured(out, orphanSummary(out)), nil
	})

	// 4. note_list_tags ----------------------------------------------------
	toolTags := mcp.NewTool("note_list_tags",
		mcp.WithDescription("List the tags used in a vault with their note counts, read from the maintained tag index (frontmatter tags/tag fields plus inline #tag, excluding code blocks, inline code and headings). Use it to see how a vault is organised before searching."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("prefix", mcp.Description("Only return tags starting with this prefix, e.g. 'project/' for nested tags. Empty returns all.")),
		mcp.WithString("scope", mcp.Description("Source filter: 'all' (default), 'frontmatter' or 'inline'.")),
		mcp.WithNumber("limit", mcp.Description("Maximum tags to return, default 200, 0 means unlimited.")),
		mcp.WithOutputSchema[dto.McpNoteTagListOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolTags, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault := resolveVaultArg(ctx, appContainer, args)
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		prefix, _ := args["prefix"].(string)
		scope, _ := args["scope"].(string)
		if scope == "" {
			scope = "all"
		}
		limit := intArg(args, "limit", 200)

		vaultID, err := appContainer.VaultService.MustGetID(ctx, uid, vault)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		tags, err := appContainer.NoteTagRepo.ListByVault(ctx, vaultID, uid, prefix, scope, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		out := &dto.McpNoteTagListOutput{
			Vault: vault,
			Scope: scope,
			Count: len(tags),
			Tags:  make([]*dto.McpTagDTO, 0, len(tags)),
		}
		for _, t := range tags {
			out.Tags = append(out.Tags, &dto.McpTagDTO{Name: t.Tag, Count: t.Count})
		}
		return mcp.NewToolResultStructured(out, tagListSummary(out)), nil
	})

	// 5. note_list_recent --------------------------------------------------
	toolRecent := mcp.NewTool("note_list_recent",
		mcp.WithDescription("List the most recently modified notes in a vault, newest first, metadata only. This is the 'what was I working on' query and costs far less than listing the whole vault."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithNumber("limit", mcp.Description("How many notes to return, default 20, max 100.")),
		mcp.WithOutputSchema[mcpNoteListOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolRecent, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault := resolveVaultArg(ctx, appContainer, args)
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		limit := clampLimit(intArg(args, "limit", defaultQueryLimit))

		pager := &pkgapp.Pager{Page: 1, PageSize: limit}
		notes, _, err := appContainer.NoteService.WithClient(getClientInfoFromContext(ctx)).List(ctx, uid, &dto.NoteListRequest{
			Vault:     vault,
			SortBy:    "mtime",
			SortOrder: "desc",
		}, pager)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		out := &mcpNoteListOutput{Vault: vault, Count: len(notes), Notes: make([]*dto.McpNoteNoContentDTO, 0, len(notes))}
		for _, n := range notes {
			out.Notes = append(out.Notes, n.ToMcpNoteNoContentDTO())
		}
		return mcp.NewToolResultStructured(out, "recent notes in "+vault+": "+itoa(out.Count)), nil
	})

	// 6. note_list_by_prefix ----------------------------------------------
	toolByPrefix := mcp.NewTool("note_list_by_prefix",
		mcp.WithDescription("List notes under a folder path prefix, metadata only, sorted by path. Use it to walk a directory structure without fetching note bodies."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("prefix", mcp.Required(), mcp.Description("Folder path prefix without a leading slash, e.g. 'projects' or '基础设施'.")),
		mcp.WithOutputSchema[mcpNoteListOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolByPrefix, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault := resolveVaultArg(ctx, appContainer, args)
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		prefix, _ := args["prefix"].(string)
		prefix = strings.Trim(prefix, "/")

		vaultID, err := appContainer.VaultService.MustGetID(ctx, uid, vault)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		notes, err := appContainer.NoteRepo.ListByPathPrefix(ctx, prefix, vaultID, uid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sort.Slice(notes, func(i, j int) bool { return notes[i].Path < notes[j].Path })

		out := &mcpNoteListOutput{Vault: vault, Count: len(notes), Notes: make([]*dto.McpNoteNoContentDTO, 0, len(notes))}
		for _, n := range notes {
			out.Notes = append(out.Notes, &dto.McpNoteNoContentDTO{
				Path:             n.Path,
				PathHash:         n.PathHash,
				Version:          n.Version,
				Size:             n.Size,
				Ctime:            n.Ctime,
				Mtime:            n.Mtime,
				ClientName:       n.ClientName,
				ClientType:       n.ClientType,
				ClientVersion:    n.ClientVersion,
				UpdatedTimestamp: n.UpdatedTimestamp,
				UpdatedAt:        formatNoteTime(n.UpdatedAt),
				CreatedAt:        formatNoteTime(n.CreatedAt),
			})
		}
		return mcp.NewToolResultStructured(out, "notes under '"+prefix+"': "+itoa(out.Count)), nil
	})

	// 7. vault_stats -------------------------------------------------------
	toolVaultStats := mcp.NewTool("vault_stats",
		mcp.WithDescription("Aggregate statistics for one vault: note and file counts and sizes, total link edges and distinct tag count. Use it to size up a vault before planning a bulk operation."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithOutputSchema[dto.McpVaultStatsOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolVaultStats, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault := resolveVaultArg(ctx, appContainer, args)
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		v, err := appContainer.VaultRepo.GetByName(ctx, vault, uid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if v == nil {
			return mcp.NewToolResultError("vault not found: " + vault), nil
		}

		out := &dto.McpVaultStatsOutput{
			Vault:     v.Name,
			NoteCount: v.NoteCount,
			NoteSize:  v.NoteSize,
			FileCount: v.FileCount,
			FileSize:  v.FileSize,
			TotalSize: v.TotalSize(),
			CreatedAt: formatNoteTime(v.CreatedAt),
			UpdatedAt: formatNoteTime(v.UpdatedAt),
		}
		if links, err := appContainer.NoteLinkRepo.ListByVault(ctx, v.ID, uid, 0); err == nil {
			out.LinkCount = len(links)
		}
		if tags, err := appContainer.NoteTagRepo.ListByVault(ctx, v.ID, uid, "", "all", 0); err == nil {
			out.TagCount = len(tags)
		}
		return mcp.NewToolResultStructured(out, vaultStatsSummary(out)), nil
	})
}

// ------------------------------------------------------------------ helpers

// resolveVaultArg applies the same "omitted or 'default'" rule the note tools use.
// resolveVaultArg 沿用笔记工具的“省略或 default”库名解析规则。
func resolveVaultArg(ctx context.Context, appContainer *app.App, args map[string]interface{}) string {
	vault, _ := args["vault"].(string)
	if vault == "" || strings.EqualFold(vault, "default") {
		vault = getDefaultVaultName(ctx, appContainer)
	}
	return vault
}

func intArg(args map[string]interface{}, key string, def int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return def
}

func pageArgs(args map[string]interface{}, defSize int) (page, pageSize int) {
	page = intArg(args, "page", 1)
	if page < 1 {
		page = 1
	}
	return page, clampLimit(intArg(args, "pageSize", defSize))
}

func clampLimit(n int) int {
	if n <= 0 {
		return defaultQueryLimit
	}
	if n > maxQueryLimit {
		return maxQueryLimit
	}
	return n
}

// itoa avoids pulling strconv into every summary line.
// itoa 避免为了拼摘要行引入 strconv。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// formatNoteTime renders time.Time the way the MCP output schemas expect
// (a string), returning "" for the zero value.
// formatNoteTime 按 MCP 输出 schema 要求把 time.Time 渲染为字符串，零值返回空串。
func formatNoteTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return timex.Time(t).String()
}

// noteBaseKey is the lowercase basename without the .md suffix, which is how
// Obsidian resolves short wikilinks like [[Note Name]].
// noteBaseKey 是小写文件名去掉 .md 后缀，对应 Obsidian 短链接 [[笔记名]] 的解析方式。
func noteBaseKey(p string) string {
	return strings.TrimSuffix(strings.ToLower(path.Base(p)), ".md")
}

// loadGraph fetches the vault's note metadata and its link edges, both without
// loading note bodies.
// loadGraph 取库内笔记元数据与链接边，两者都不加载正文。
func loadGraph(ctx context.Context, appContainer *app.App, vault string, uid int64) ([]*domain.Note, []*domain.NoteLink, error) {
	vaultID, err := appContainer.VaultService.MustGetID(ctx, uid, vault)
	if err != nil {
		return nil, nil, err
	}
	notes, err := appContainer.NoteRepo.ListMetaByVault(ctx, vaultID, uid)
	if err != nil {
		return nil, nil, err
	}
	links, err := appContainer.NoteLinkRepo.ListByVault(ctx, vaultID, uid, 0)
	if err != nil {
		return nil, nil, err
	}
	return notes, links, nil
}

// buildNoteIndexes indexes notes by ID, by full path hash, and by lowercase
// basename so short wikilinks resolve the way Obsidian resolves them.
// buildNoteIndexes 按 ID、完整路径哈希、小写文件名三种方式索引笔记，
// 使短双链能像 Obsidian 一样解析。
func buildNoteIndexes(notes []*domain.Note) (map[int64]string, map[string]string, map[string]string) {
	pathByID := make(map[int64]string, len(notes))
	byHash := make(map[string]string, len(notes))
	byBase := make(map[string]string, len(notes))
	for _, n := range notes {
		pathByID[n.ID] = n.Path
		if n.PathHash != "" {
			byHash[n.PathHash] = n.Path
		}
		if k := noteBaseKey(n.Path); k != "" {
			byBase[k] = n.Path
		}
	}
	return pathByID, byHash, byBase
}

// resolveLinkPath maps a wikilink target to a vault path, trying the stored
// hash, then the hash of the raw target text, then the basename. A link that
// resolves by basename is reported as intact even when it does not carry the
// folder prefix — same tolerance Obsidian applies.
// resolveLinkPath 把双链目标映射到库内路径：先试存储的哈希，再试原始文本的哈希，
// 最后试文件名。仅靠文件名命中的链接也算完好（不带目录前缀）——与 Obsidian 的宽容度一致。
func resolveLinkPath(targetPath, targetPathHash string, byHash, byBase map[string]string) (string, bool) {
	if targetPathHash != "" {
		if p, ok := byHash[targetPathHash]; ok {
			return p, true
		}
	}
	if targetPath != "" {
		if p, ok := byHash[util.EncodeHash32(targetPath)]; ok {
			return p, true
		}
		if p, ok := byBase[noteBaseKey(targetPath)]; ok {
			return p, true
		}
	}
	return "", false
}

// matchLineOf returns the 1-based line number of the first case-insensitive
// match, or 0 when the keyword is not present (path-only matches).
// matchLineOf 返回首个忽略大小写命中的行号（从 1 起），未命中返回 0（仅路径命中）。
func matchLineOf(content, keyword string) int {
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if kw == "" {
		return 0
	}
	for i, line := range strings.Split(content, "\n") {
		if strings.Contains(strings.ToLower(line), kw) {
			return i + 1
		}
	}
	return 0
}

// ------------------------------------------------------------ text summaries
// The structured payload is the contract; these one-liners make the
// content-companion readable in clients that only render text.

func searchSummary(o *dto.McpNoteSearchOutput) string {
	s := o.Keyword + " → " + itoa(o.Count) + "/" + itoa(o.Total) + " hits (" + o.Mode + ") in " + o.Vault
	if o.HasMore {
		s += ", page " + itoa(o.Page) + " has more"
	}
	return s
}

func brokenLinksSummary(o *dto.McpNoteBrokenLinksOutput) string {
	return itoa(o.Count) + " broken of " + itoa(o.TotalLinks) + " links in " + o.Vault
}

func orphanSummary(o *dto.McpNoteOrphanListOutput) string {
	return itoa(o.Count) + " orphans of " + itoa(o.TotalNotes) + " notes in " + o.Vault
}

func tagListSummary(o *dto.McpNoteTagListOutput) string {
	return itoa(o.Count) + " tags in " + o.Vault + " (scope " + o.Scope + ")"
}

func vaultStatsSummary(o *dto.McpVaultStatsOutput) string {
	return o.Vault + ": " + itoa(int(o.NoteCount)) + " notes, " + itoa(int(o.LinkCount)) + " links, " + itoa(o.TagCount) + " tags"
}
