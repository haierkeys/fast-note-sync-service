package diff

import (
	"strings"
	"testing"
)

// 多设备同步真实场景测试
// 模拟 Obsidian 笔记在多设备间同步时可能遇到的各种情况

// TestScenario_DailyNoteEditing 日记编辑场景
// 用户在手机和电脑上同时编辑同一篇日记
func TestScenario_DailyNoteEditing(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		phone        string // 手机端修改
		desktop      string // 电脑端修改
		wantConflict bool
		description  string
	}{
		{
			name: "不同段落编辑-无冲突",
			base: `# 2024-01-10 日记

## 早晨
今天早起跑步了

## 下午
下午开会讨论项目

## 晚上
晚上看了一部电影`,
			phone: `# 2024-01-10 日记

## 早晨
今天早起跑步了，跑了5公里

## 下午
下午开会讨论项目

## 晚上
晚上看了一部电影`,
			desktop: `# 2024-01-10 日记

## 早晨
今天早起跑步了

## 下午
下午开会讨论项目，决定下周上线

## 晚上
晚上看了一部电影`,
			wantConflict: false,
			description:  "手机编辑早晨部分，电脑编辑下午部分，应该能自动合并",
		},
		{
			name: "同一段落编辑-冲突",
			base: `# 会议记录

参会人员：张三、李四
会议内容：讨论Q1计划`,
			phone: `# 会议记录

参会人员：张三、李四、王五
会议内容：讨论Q1计划`,
			desktop: `# 会议记录

参会人员：张三、李四、赵六
会议内容：讨论Q1计划`,
			wantConflict: true,
			description:  "两端都修改了参会人员列表，应该检测到冲突",
		},
		{
			name: "一端删除段落-另一端修改同段落-冲突",
			base: `# 待办事项

- [ ] 买菜
- [ ] 取快递
- [ ] 交水电费`,
			phone: `# 待办事项

- [ ] 买菜
- [ ] 交水电费`,
			desktop: `# 待办事项

- [ ] 买菜
- [x] 取快递
- [ ] 交水电费`,
			wantConflict: true,
			description:  "手机删除了取快递，电脑标记取快递完成，应该冲突",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTexts(tt.base, tt.phone, tt.desktop, true)
			if err != nil {
				t.Fatalf("MergeTexts() error = %v", err)
			}

			if result.HasConflict != tt.wantConflict {
				t.Errorf("%s\nHasConflict = %v, want %v\nConflictInfo: %s",
					tt.description, result.HasConflict, tt.wantConflict, result.ConflictInfo)
			}
		})
	}
}

// TestScenario_CodeSnippetEditing 代码片段编辑场景
func TestScenario_CodeSnippetEditing(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		device1      string
		device2      string
		wantConflict bool
		description  string
	}{
		{
			name: "添加不同函数-无冲突",
			base: "```go\npackage main\n\nfunc main() {\n}\n```",
			device1: "```go\npackage main\n\nfunc hello() {\n\tprintln(\"hello\")\n}\n\nfunc main() {\n}\n```",
			device2: "```go\npackage main\n\nfunc main() {\n\thello()\n}\n```",
			wantConflict: false,
			description:  "一端添加函数定义，另一端修改main函数，应该能合并",
		},
		{
			name: "修改同一行代码-冲突",
			base: "```python\ndef calculate(x):\n    return x * 2\n```",
			device1: "```python\ndef calculate(x):\n    return x * 3\n```",
			device2: "```python\ndef calculate(x):\n    return x + 2\n```",
			wantConflict: true,
			description:  "两端都修改了计算逻辑，应该冲突",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTexts(tt.base, tt.device1, tt.device2, true)
			if err != nil {
				t.Fatalf("MergeTexts() error = %v", err)
			}

			if result.HasConflict != tt.wantConflict {
				t.Errorf("%s\nHasConflict = %v, want %v", tt.description, result.HasConflict, tt.wantConflict)
			}
		})
	}
}

// TestScenario_ListEditing 列表编辑场景
func TestScenario_ListEditing(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		device1      string
		device2      string
		wantConflict bool
		wantContains []string // 合并后应该包含的内容
		description  string
	}{
		{
			name:         "列表末尾添加不同项-无冲突",
			base:         "- 苹果\n- 香蕉",
			device1:      "- 苹果\n- 香蕉\n- 橙子",
			device2:      "- 苹果\n- 香蕉\n- 葡萄",
			wantConflict: false,
			wantContains: []string{"苹果", "香蕉", "橙子", "葡萄"},
			description:  "两端都在列表末尾添加项目，应该都保留",
		},
		{
			name:         "列表开头添加不同项-无冲突",
			base:         "- 项目B\n- 项目C",
			device1:      "- 项目A\n- 项目B\n- 项目C",
			device2:      "- 项目B\n- 项目C\n- 项目D",
			wantConflict: false,
			wantContains: []string{"项目A", "项目B", "项目C", "项目D"},
			description:  "一端在开头添加，一端在末尾添加，应该都保留",
		},
		{
			name:         "删除同一项-无冲突",
			base:         "- 保留1\n- 删除我\n- 保留2",
			device1:      "- 保留1\n- 保留2",
			device2:      "- 保留1\n- 保留2",
			wantConflict: false,
			wantContains: []string{"保留1", "保留2"},
			description:  "两端都删除同一项，应该成功合并",
		},
		{
			name:         "修改同一列表项-冲突",
			base:         "- 任务：写文档",
			device1:      "- 任务：写文档（已完成）",
			device2:      "- 任务：写文档（进行中）",
			wantConflict: true,
			description:  "两端对同一任务状态做了不同修改",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTexts(tt.base, tt.device1, tt.device2, true)
			if err != nil {
				t.Fatalf("MergeTexts() error = %v", err)
			}

			if result.HasConflict != tt.wantConflict {
				t.Errorf("%s\nHasConflict = %v, want %v", tt.description, result.HasConflict, tt.wantConflict)
			}

			if !result.HasConflict && len(tt.wantContains) > 0 {
				for _, want := range tt.wantContains {
					if !strings.Contains(result.Content, want) {
						t.Errorf("合并结果应该包含 %q，但没有找到\n结果: %s", want, result.Content)
					}
				}
			}
		})
	}
}

// TestScenario_OfflineEditing 离线编辑场景
// 模拟设备离线后重新上线同步的情况
func TestScenario_OfflineEditing(t *testing.T) {
	tests := []struct {
		name         string
		base         string // 离线前的共同版本
		offline      string // 离线设备的修改
		online       string // 在线设备的修改
		wantConflict bool
		description  string
	}{
		{
			name: "离线添加内容-在线也添加内容-不同位置",
			base: `# 读书笔记

第一章要点：
- 要点1`,
			offline: `# 读书笔记

第一章要点：
- 要点1
- 要点2（离线添加）`,
			online: `# 读书笔记

作者：某某某

第一章要点：
- 要点1`,
			wantConflict: false,
			description:  "离线设备添加要点，在线设备添加作者信息，应该能合并",
		},
		{
			name: "离线删除段落-在线修改同段落-冲突",
			base: `# 项目计划

## 第一阶段
需求分析

## 第二阶段
开发实现`,
			offline: `# 项目计划

## 第二阶段
开发实现`,
			online: `# 项目计划

## 第一阶段
需求分析和设计

## 第二阶段
开发实现`,
			wantConflict: true,
			description:  "离线设备删除了第一阶段，在线设备修改了第一阶段内容",
		},
		{
			name: "长时间离线-大量修改-不同区域",
			base: `# 工作周报

## 本周完成
- 任务A

## 下周计划
- 任务B

## 问题与风险
无`,
			offline: `# 工作周报

## 本周完成
- 任务A
- 任务C（离线添加）
- 任务D（离线添加）

## 下周计划
- 任务B

## 问题与风险
无`,
			online: `# 工作周报

## 本周完成
- 任务A

## 下周计划
- 任务B
- 任务E（在线添加）

## 问题与风险
发现一个技术难点`,
			wantConflict: false,
			description:  "离线和在线分别在不同区域做了大量修改，应该能合并",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTexts(tt.base, tt.offline, tt.online, true)
			if err != nil {
				t.Fatalf("MergeTexts() error = %v", err)
			}

			if result.HasConflict != tt.wantConflict {
				t.Errorf("%s\nHasConflict = %v, want %v\nConflictInfo: %s",
					tt.description, result.HasConflict, tt.wantConflict, result.ConflictInfo)
			}
		})
	}
}

// TestScenario_FrontmatterEditing YAML Frontmatter 编辑场景
func TestScenario_FrontmatterEditing(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		device1      string
		device2      string
		wantConflict bool
		description  string
	}{
		{
			name: "修改不同属性-无冲突",
			base: `---
title: 我的笔记
tags: [日记]
date: 2024-01-10
---

正文内容`,
			device1: `---
title: 我的笔记（已更新）
tags: [日记]
date: 2024-01-10
---

正文内容`,
			device2: `---
title: 我的笔记
tags: [日记, 重要]
date: 2024-01-10
---

正文内容`,
			wantConflict: false,
			description:  "一端修改标题，一端修改标签，应该能合并",
		},
		{
			name: "修改同一属性-冲突",
			base: `---
status: draft
---

内容`,
			device1: `---
status: published
---

内容`,
			device2: `---
status: archived
---

内容`,
			wantConflict: true,
			description:  "两端都修改了status属性，应该冲突",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTexts(tt.base, tt.device1, tt.device2, true)
			if err != nil {
				t.Fatalf("MergeTexts() error = %v", err)
			}

			if result.HasConflict != tt.wantConflict {
				t.Errorf("%s\nHasConflict = %v, want %v", tt.description, result.HasConflict, tt.wantConflict)
			}
		})
	}
}

// TestScenario_TableEditing 表格编辑场景
func TestScenario_TableEditing(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		device1      string
		device2      string
		wantConflict bool
		description  string
	}{
		{
			name: "添加不同行-一端末尾一端中间-冲突",
			base: `| 姓名 | 分数 |
|------|------|
| 张三 | 90   |`,
			device1: `| 姓名 | 分数 |
|------|------|
| 张三 | 90   |
| 李四 | 85   |`,
			device2: `| 姓名 | 分数 |
|------|------|
| 王五 | 88   |
| 张三 | 90   |`,
			wantConflict: true,
			description:  "一端在末尾添加行，一端在中间插入行，由于 diff 算法限制视为冲突",
		},
		{
			name: "修改同一单元格-冲突",
			base: `| 项目 | 状态 |
|------|------|
| 功能A | 进行中 |`,
			device1: `| 项目 | 状态 |
|------|------|
| 功能A | 已完成 |`,
			device2: `| 项目 | 状态 |
|------|------|
| 功能A | 已取消 |`,
			wantConflict: true,
			description:  "两端都修改了功能A的状态",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTexts(tt.base, tt.device1, tt.device2, true)
			if err != nil {
				t.Fatalf("MergeTexts() error = %v", err)
			}

			if result.HasConflict != tt.wantConflict {
				t.Errorf("%s\nHasConflict = %v, want %v", tt.description, result.HasConflict, tt.wantConflict)
			}
		})
	}
}

// TestScenario_LinkEditing 链接编辑场景
func TestScenario_LinkEditing(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		device1      string
		device2      string
		wantConflict bool
		description  string
	}{
		{
			name:         "添加不同链接-无冲突",
			base:         "相关笔记：\n- [[笔记A]]",
			device1:      "相关笔记：\n- [[笔记A]]\n- [[笔记B]]",
			device2:      "相关笔记：\n- [[笔记A]]\n- [[笔记C]]",
			wantConflict: false,
			description:  "两端添加不同的链接",
		},
		{
			name:         "修改同一链接-冲突",
			base:         "参考：[[旧链接]]",
			device1:      "参考：[[新链接1]]",
			device2:      "参考：[[新链接2]]",
			wantConflict: true,
			description:  "两端都修改了同一个链接",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTexts(tt.base, tt.device1, tt.device2, true)
			if err != nil {
				t.Fatalf("MergeTexts() error = %v", err)
			}

			if result.HasConflict != tt.wantConflict {
				t.Errorf("%s\nHasConflict = %v, want %v", tt.description, result.HasConflict, tt.wantConflict)
			}
		})
	}
}

// TestScenario_EmptyAndWhitespace 空内容和空白字符场景
func TestScenario_EmptyAndWhitespace(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		device1      string
		device2      string
		wantConflict bool
		description  string
	}{
		{
			name:         "从空文件开始-两端都添加内容",
			base:         "",
			device1:      "设备1添加的内容",
			device2:      "设备2添加的内容",
			wantConflict: true,
			description:  "空文件两端同时添加不同内容，应该冲突",
		},
		{
			name:         "一端清空文件-另一端修改-冲突",
			base:         "原始内容\n第二行",
			device1:      "",
			device2:      "原始内容（已修改）\n第二行",
			wantConflict: true,
			description:  "一端清空文件，另一端修改内容",
		},
		{
			name:         "两端都清空文件-无冲突",
			base:         "要删除的内容",
			device1:      "",
			device2:      "",
			wantConflict: false,
			description:  "两端都清空文件，结果一致",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTexts(tt.base, tt.device1, tt.device2, true)
			if err != nil {
				t.Fatalf("MergeTexts() error = %v", err)
			}

			if result.HasConflict != tt.wantConflict {
				t.Errorf("%s\nHasConflict = %v, want %v\nConflictInfo: %s",
					tt.description, result.HasConflict, tt.wantConflict, result.ConflictInfo)
			}
		})
	}
}
