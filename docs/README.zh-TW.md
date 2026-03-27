[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

如有問題請建立 [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new)，或加入 Telegram 交流群尋求協助：[https://t.me/obsidian_users](https://t.me/obsidian_users)

中國大陸地區，推薦使用騰訊 `cnb.cool` 鏡像庫：[https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)



<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高效能、低延遲的筆記同步、線上管理與遠端 REST API 服務平台</strong>
  <br>
  <em>基於 Golang + WebSocket + SQLite + React 構建</em>
</p>

<p align="center">
  資料提供需搭配客戶端插件使用：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
  <div align="center">
    <a href="/docs/images/vault.png"><img src="/docs/images/vault.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="/docs/images/attach.png"><img src="/docs/images/attach.png" alt="fast-note-sync-service-preview" width="400" /></a>
    </div>
  <div align="center">
    <a href="/docs/images/note.png"><img src="/docs/images/note.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="/docs/images/setting.png"><img src="/docs/images/setting.png" alt="fast-note-sync-service-preview" width="400" /></a>
  </div>
</div>

---

## ✨ 核心功能

* **🧰 MCP (Model Context Protocol) 原生支援**：
  * `FNS` 可以作為 MCP 服務端接入 `Cherry Studio`、`Cursor` 等相容的 AI 客戶端，即可讓 AI 具備讀寫私人筆記與附件的能力，且所有變更會透過 WebSocket 即時同步至各裝置終端。
* **🚀 REST API 支援**：
    * 提供標準的 REST API 介面，支援透過程式化方式（如自動化腳本、AI 助手整合）對 Obsidian 筆記進行增刪改查。
    * 詳情請參閱 [RESTful API 文件](/docs/REST_API.md) 或 [OpenAPI 文件](/docs/swagger.yaml)。
* **💻 Web 管理面板**：
  * 內建現代化管理介面，輕鬆建立使用者、產生插件設定、管理庫及筆記內容。
* **🔄 多端筆記同步**：
    * 支援 **Vault（倉庫）** 自動建立。
    * 支援筆記管理（增、刪、改、查），變更毫秒級即時分發至所有在線裝置。
* **🖼️ 附件同步支援**：
    * 完美支援圖片等非筆記檔案同步。
    * 支援大附件分片上傳下載，分片大小可設定，提升同步效率。
* **⚙️ 設定同步**：
    * 支援 `.obsidian` 設定檔的同步。
    * 支援 `PDF` 進度狀態同步。
* **📝 筆記歷史**：
    * 可在 Web 頁面、插件端查看每個筆記的歷史修改版本。
    * （需服務端 v1.2+）
* **🗑️ 資源回收桶**：
    * 支援筆記刪除後，自動進入資源回收桶。
    * 支援從資源回收桶恢復筆記。（後續將陸續新增附件恢復功能）

* **🚫 離線同步策略**：
    * 支援筆記離線編輯自動合併。（需插件端設定）
    * 離線刪除，重新連線後自動補全或刪除同步。（需插件端設定）

* **🔗 分享功能**：
    * 可以建立/取消筆記分享。
    * 自動解析分享筆記中引用的圖片、音視頻等附件。
    * 提供分享存取統計功能。
    * 可以設定分享筆記的訪問密碼。
    * 可以對分享筆記產生短連結。
* **📂 目錄同步**：
    * 支援資料夾的建立/重新命名/移動/刪除同步。

* **🌳 Git 自動化**：
    * 當附件和筆記發生變更時，自動更新並推送至遠端 Git 倉庫。
    * 任務結束後自動釋放系統記憶體。

* **☁️ 多儲存備份與單向鏡像同步**：
    * 適配 S3/OSS/R2/WebDAV/本地等多種儲存協議。
    * 支援全量/增量 ZIP 定時歸檔備份。
    * 支援 Vault 資源單向鏡像同步至遠端儲存。
    * 自動清理過期備份，支援自訂保留天數。

## ☕ 贊助與支持

- 如果您覺得這個專案很有用，並希望它繼續開發，請透過以下方式支持我：

  | Ko-fi *（非中國地區）*                                                                            |    | 微信掃碼打賞 *（中國地區）*                    |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | 或 | <img src="/docs/images/wxds.png" height="150"> |

  - 已支持名單：
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md (cnb.cool 鏡像庫)</a>

## ⏱️ 更新日誌

- ♨️ [查看更新日誌](/docs/CHANGELOG.zh-TW.md)

## 🗺️ 路線圖 (Roadmap)

我們正在持續改進，以下是未來的開發計劃：


- [ ] **更多資料庫類型的支援**


> **如果您有改進建議或新想法，歡迎透過提交 issue 與我們分享——我們會認真評估並採納合適的建議。**

## 🚀 快速部署

我們提供多種安裝方式，推薦使用 **一鍵腳本** 或 **Docker**。

### 方式一：一鍵腳本（推薦）

自動偵測系統環境並完成安裝、服務登錄。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

中國地區可以使用騰訊 `cnb.cool` 鏡像源：
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**腳本主要行為：**

  * 自動下載適配當前系統的 Release 二進位檔案。
  * 預設安裝至 `/opt/fast-note`，並在 `/usr/local/bin/fns` 建立全域捷徑命令 `fns`。
  * 設定並啟動 Systemd（Linux）或 Launchd（macOS）服務，實現開機自啟。
  * **管理命令**：`fns [install|uninstall|start|stop|status|update|menu]`
  * **互動選單**：直接執行 `fns` 可進入互動選單，支援安裝/升級、服務控制、開機自啟設定，以及在 GitHub / CNB 鏡像之間切換。

-----

### 方式二：Docker 部署

#### Docker Run

```bash
# 1. 拉取鏡像
docker pull haierkeys/fast-note-sync-service:latest

# 2. 啟動容器
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

建立 `docker-compose.yaml` 檔案：

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # RESTful API & WebSocket 埠；/api/user/sync 為 WebSocket 介面位址
    volumes:
      - ./storage:/fast-note-sync/storage  # 資料儲存
      - ./config:/fast-note-sync/config    # 設定檔
```

啟動服務：

```bash
docker compose up -d
```

-----

### 方式三：手動二進位安裝

從 [Releases](https://github.com/haierkeys/fast-note-sync-service/releases) 下載對應系統的最新版本，解壓後執行：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用指南

1.  **存取管理面板**：
    在瀏覽器開啟 `http://{伺服器IP}:9000`。
2.  **初始化設定**：
    首次存取需註冊帳號。*（如需關閉註冊功能，請在設定檔中設定 `user.register-is-enable: false`）*
3.  **設定客戶端**：
    登入管理面板，點擊 **「複製 API 設定」**。
4.  **連接 Obsidian**：
    開啟 Obsidian 插件設定頁面，貼上剛才複製的設定資訊即可。


## ⚙️ 設定說明

預設設定檔為 `config.yaml`，程式會自動在 **根目錄** 或 **config/** 目錄下查找。

查看完整設定範例：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx 反向代理設定範例

查看完整設定範例：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP（模型上下文協議）支援

FNS 原生支援 **MCP (Model Context Protocol)**。

`FNS` 可以作為 MCP 服務端接入 `Cherry Studio`、`Cursor` 等相容的 AI 客戶端，即可讓 AI 具備讀寫私人筆記與附件的能力，且所有變更會透過 WebSocket 即時同步至各裝置終端。

### 接入設定（SSE 模式）

FNS 透過 **SSE 協議**提供 MCP 介面，通用參數要求如下：
- **介面位址**：`http://<您的伺服器IP或網域>:<埠>/api/mcp/sse`
- **驗證 Header**：`token: <您的 API Token>`（在 WebGUI 的複製 API 設定中取得）


#### 範例：Cherry Studio

請在 Cherry Studio 的 MCP Server 設定中新增如下 JSON 節點。
*（注：請將 `<ServerIP>`、`<Port>` 和 `<Token>` 替換為您自己的實際資訊）*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp/sse",
      "type": "sse",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer <Token>"
      }
    }
  }
}
```

## 🔗 客戶端 & 客戶端插件

* Obsidian Fast Note Sync 插件
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool 鏡像庫](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* 第三方客戶端
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) — 基於 Python 和 FNS WS 介面實現的雙向即時同步命令列客戶端，適用於無 GUI 的 Linux 伺服器環境（如 OpenClaw），實現與 Obsidian 桌面/行動端同等的同步能力。
