[中文文档](readme-zh.md) / [English Document](README.md)

<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能、低延迟的笔记同步服务解决方案</strong>
  <br>
  <em>基于 Golang + Websocket + Sqlite + React 构建</em>
</p>

<p align="center">
  需配合客户端插件使用：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ 核心功能

* **💻 Web 管理面板**：内置现代化管理界面，轻松创建用户、生成插件配置、管理仓库及笔记内容。
* **🔄 多端实时同步**：
    * 支持 **Vault (仓库)** 自动创建。
    * 支持笔记管理（增、删、改、查），变更毫秒级实时分发至所有在线设备。
* **🖼️ 附件同步支持**：
    * 完美支持图片等非笔记文件同步。
    * *(注：需服务端 v0.9+ 及 Obsidian 插件端 v1.0+, 不支持 Obsidian 设置文件)*

## 🗺️ 路线图 (Roadmap)

我们正在持续改进，以下是未来的开发计划：

- [ ] **Git 版本控制集成**：为笔记提供更安全的版本回溯。
- [ ] **同步算法优化**：集成 `google-diff-match-patch` 以实现更高效的增量同步。
- [ ] **云存储与备份策略**：
    - [ ] 自定义备份策略配置。
    - [ ] 多协议适配：S3 / Minio / Cloudflare R2 / 阿里云 OSS / WebDAV。

## 🚀 快速部署

我们提供多种安装方式，推荐使用 **一键脚本** 或 **Docker**。

### 方式一：一键脚本（推荐）

自动检测系统环境并完成安装、服务注册。

```bash
bash <(curl -fsSL [https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/quest_install.sh](https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/quest_install.sh))
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

查看完整配置示例：[config/config.yaml](https://www.google.com/search?q=config/config.yaml)

## 📅 更新日志

查看完整的版本迭代记录，请访问 [Releases 页面](https://github.com/haierkeys/fast-note-sync-service/releases)。

## ☕ 赞助与支持

本项目完全开源免费。如果您觉得它对您有帮助，欢迎 **Star** 本项目，或请作者喝一杯咖啡，这将成为我持续维护的动力，感谢！

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 🔗 相关资源

  * [Obsidian Fast Note Sync Plugin (客户端插件)](https://github.com/haierkeys/obsidian-fast-note-sync)
