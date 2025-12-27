[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能、低延遲的筆記同步服務解決方案</strong>
  <br>
  <em>基於 Golang + Websocket + Sqlite + React 構建</em>
</p>

<p align="center">
  需配合客戶端插件使用：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ 核心功能

* **💻 Web 管理面板**：內建現代化管理介面，輕鬆創建用戶、生成插件配置、管理倉庫及筆記內容。
* **🔄 多端筆記同步**：
    * 支持 **Vault (倉庫)** 自動創建。
    * 支持筆記管理（增、刪、改、查），變更毫秒級實時分發至所有在線設備。
* **🖼️ 附件同步支持**：
    * 完美支持圖片等非筆記文件同步。
    * *(註：需服務端 v0.9+ 及 [Obsidian 插件端 v1.0+ ](https://github.com/haierkeys/obsidian-fast-note-sync/releases), 不支持 Obsidian 設置文件)*
* **📝 筆記歷史**：
    * 可以在 Web 頁面，插件端查看每一個筆記的 歷史修改版本。
    * (需服務端 v1.2+ )
* **⚙️ 配置同步**：
    * 支持 `.obsidian` 配置文件的同步。


## 🗺️ 路線圖 (Roadmap)

我們正在持續改進，以下是未來的開發計劃：

- [ ] **Git 版本控制集成**：為筆記提供更安全的版本回溯。
- [ ] **同步算法優化**：集成 `google-diff-match-patch` 以實現更高效的增量同步。
- [ ] **雲存儲與備份策略**：
    * [ ] 自定義備份策略配置。
    * [ ] 多協議適配：S3 / Minio / Cloudflare R2 / 阿里雲 OSS / WebDAV。

> **如果您有改進建議或新想法，歡迎通過提交 issue 與我們分享——我們會認真評估並採納合適的建議。**

## 🚀 快速部署

我們提供多種安裝方式，推薦使用 **一鍵腳本** 或 **Docker**。

### 方式一：一鍵腳本（推薦）

自動檢測系統環境並完成安裝、服務註冊。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**腳本主要行為：**

  * 自動下載適配當前系統的 Release 二進制文件。
  * 默認安裝至 `/opt/fast-note`，並在 `/usr/local/bin/fast-note` 創建快捷指令。
  * 配置並啟動 Systemd 服務 (`fast-note.service`)，實現開機自啟。
  * **管理命令**：`fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### 方式二：Docker 部署

#### Docker Run

```bash
# 1. 拉取鏡像
docker pull haierkeys/fast-note-sync-service:latest

# 2. 啟動容器
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 -p 9001:9001 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

創建 `docker-compose.yaml` 文件：

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
      - ./storage:/fast-note-sync/storage  # 數據存儲
      - ./config:/fast-note-sync/config    # 配置文件
```

啟動服務：

```bash
docker compose up -d
```

-----

### 方式三：手動二進制安裝

從 [Releases](https://github.com/haierkeys/fast-note-sync-service/releases) 下載對應系統的最新版本，解壓後運行：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用指南

1.  **訪問管理面板**：
    在瀏覽器打開 `http://{服務器IP}:9000`。
2.  **初始化設置**：
    首次訪問需註冊賬號。*(如需關閉註冊功能，請在配置文件中設置 `user.register-is-enable: false`)*
3.  **配置客戶端**：
    登錄管理面板，點擊 **“複製 API 配置”**。
4.  **連接 Obsidian**：
    打開 Obsidian 插件設置頁面，粘貼剛才複製的配置信息即可。

## ⚙️ 配置說明

默認配置文件為 `config.yaml`，程序會自動在 **根目錄** 或 **config/** 目錄下查找。

查看完整配置示例：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 📅 更新日誌

查看完整的版本迭代記錄，請訪問 [Releases 頁面](https://github.com/haierkeys/fast-note-sync-service/releases)。

## ☕ 贊助與支持

本項目完全開源免費。如果您覺得它對您有幫助，歡迎 **Star** 本項目，或請作者喝一杯咖啡，這將成為我持續維護的動力，感謝！

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 🔗 相關資源

  * [Obsidian Fast Note Sync Plugin (客戶端插件)](https://github.com/haierkeys/obsidian-fast-note-sync)
