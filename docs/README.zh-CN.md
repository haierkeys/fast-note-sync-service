[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

有问题请新建 [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) , 或加入电报交流群寻求帮助: [https://t.me/obsidian_users](https://t.me/obsidian_users)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能、低延迟的笔记同步, 在线管理, 远端 REST API 服务平台</strong>
  <br>
  <em>基于 Golang + Websocket + Sqlite + React 构建</em>
</p>

<p align="center">
  数据提供需配合客户端插件使用：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
  <div align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/vault.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/vault.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/attach.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/attach.png" alt="fast-note-sync-service-preview" width="400" /></a>
    </div>
  <div align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/note.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/note.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/setting.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/setting.png" alt="fast-note-sync-service-preview" width="400" /></a>
  </div>
</div>

---

## ✨ 核心功能

* **🚀 REST API 支持**：
    * 提供标准的 REST API 接口，支持通过编程方式（如自动化脚本、AI 助手集成）对 Obsidian 笔记进行增删改查。
    * 详情请参阅 [REST API 文档](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/REST_API.md)。
* **💻 Web 管理面板**：
  * 内置现代化管理界面，轻松创建用户、生成插件配置、管理仓库及笔记内容。
* **🔄 多端笔记同步**：
    * 支持 **Vault (仓库)** 自动创建。
    * 支持笔记管理（增、删、改、查），变更毫秒级实时分发至所有在线设备。
* **🖼️ 附件同步支持**：
    * 完美支持图片等非笔记文件同步。
    * 支持大附件 分片上传下载，分片大小可配置，提升同步效率。
* **⚙️ 配置同步**：
    * 支持 `.obsidian` 配置文件的同步。
    * 支持 `PDF` 进度状态同步。
* **📝 笔记历史**：
    * 可以在 Web 页面，插件端查看每一个笔记的 历史修改版本。
    * (需服务端 v1.2+ )
* **🗑️ 回收站**：
    * 支持笔记删除后，自动进入回收站。
    * 支持从回收站恢复笔记。(后续会陆续新增附件恢复功能)

* **🚫 离线同步策略**：
    * 支持笔记离线编辑自动合并。(需要插件端设置)
    * 离线删除，重连之后自动补全或删除同步。(需要插件端设置)

## ☕ 赞助与支持

- 如果觉得这个插件很有用，并且想要它继续开发，请在以下方式支持我:

  | Ko-fi *非中国地区*                                                                                                   |    | 微信扫码打赏 *中国地区*                                            |
  |----------------------------------------------------------------------------------------------------------------------|----|--------------------------------------------------------------------|
  | [<img src="https://ik.imagekit.io/haierkeys/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | 或 | <img src="https://ik.imagekit.io/haierkeys/wxds.png" height="150"> |

  - 已支持名单：
    - https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md

## ⏱️ 更新日志

- ♨️ [访问查看更新日志](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/CHANGELOG.zh-CN.md)

## 🗺️ 路线图 (Roadmap)

我们正在持续改进，以下是未来的开发计划：

- [ ] **分享功能**：支持笔记的分享。
- [ ] **MCP支持**：增加 AI MCP 相关功能支持。
- [ ] **目录同步**：支持目录的增删改查。
- [ ] **Git 版本控制集成**：为笔记提供更安全的版本回溯。
- [ ] **云存储与备份策略**：
    - [ ] 自定义备份策略配置。
    - [ ] 多协议适配：S3 / Minio / Cloudflare R2 / 阿里云 OSS / WebDAV。

> **如果您有改进建议或新想法，欢迎通过提交 issue 与我们分享——我们会认真评估并采纳合适的建议。**

## 🚀 快速部署

我们提供多种安装方式，推荐使用 **一键脚本** 或 **Docker**。

### 方式一：一键脚本（推荐）

自动检测系统环境并完成安装、服务注册。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**脚本主要行为：**

  * 自动下载适配当前系统的 Release 二进制文件。
  * 默认安装至 `/opt/fast-note`，并在 `/usr/local/bin/fast-note` 创建快捷指令。
  * 配置并启动 Systemd 服务 (`fast-note.service`)，实现开机自启。
  * **管理命令**：`fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### 方式二：Docker 部署

#### Docker Run

```bash
# 1. 拉取镜像
docker pull haierkeys/fast-note-sync-service:latest

# 2. 启动容器
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 -p 9001:9001 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

创建 `docker-compose.yaml` 文件：

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # API 端口
      - "9001:9001"  # WebSocket 端口
    volumes:
      - ./storage:/fast-note-sync/storage  # 数据存储
      - ./config:/fast-note-sync/config    # 配置文件
```

启动服务：

```bash
docker compose up -d
```

-----

### 方式三：手动二进制安装

从 [Releases](https://github.com/haierkeys/fast-note-sync-service/releases) 下载对应系统的最新版本，解压后运行：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用指南

1.  **访问管理面板**：
    在浏览器打开 `http://{服务器IP}:9000`。
2.  **初始化设置**：
    首次访问需注册账号。*(如需关闭注册功能，请在配置文件中设置 `user.register-is-enable: false`)*
3.  **配置客户端**：
    登录管理面板，点击 **“复制 API 配置”**。
4.  **连接 Obsidian**：
    打开 Obsidian 插件设置页面，粘贴刚才复制的配置信息即可。


## ⚙️ 配置说明

默认配置文件为 `config.yaml`，程序会自动在 **根目录** 或 **config/** 目录下查找。

查看完整配置示例：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx 反代配置示例

查看完整配置示例：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 相关资源

  * [Obsidian Fast Note Sync Plugin (客户端插件)](https://github.com/haierkeys/obsidian-fast-note-sync)


