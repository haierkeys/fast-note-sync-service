[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

有問題請新建 [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) , 或加入電報交流群尋求幫助: [https://t.me/obsidian_users](https://t.me/obsidian_users)

中國大陸地區，推薦使用騰訊 cnb.cool 鏡像庫: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能、低延遲的筆記同步，線上管理，遠端 REST API 服務平台</strong>
  <br>
  <em>基於 Golang + Websocket + React 建構</em>
</p>

<p align="center">
  數據提供需配合用戶端外掛程式使用：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## 🎯 核心功能

* **🧰 MCP (Model Context Protocol) 原生支援**：
  * `FNS` 可以作為 MCP 服務端接入 `Cherry Studio`、`Cursor` 等相容的 AI 用戶端，即可讓 AI 具備讀寫私人筆記與附件的能力，且所有變更即時同步到各端。
* **🚀 REST API 支援**：
  * 提供標準的 REST API 介面，支援透過程式化方式（如自動化指令碼、AI 助手整合）對 Obsidian 筆記進行增刪改查。
  * 詳情請參閱 [RESTful API 文件](/docs/REST_API.md) 或 [OpenAPI 文件](/docs/swagger.yaml)。
* **💻 Web 管理面板**：
  * 內建現代化管理介面，輕鬆建立使用者、產生外掛組態、管理倉庫及筆記內容。
* **🔄 多端筆記同步**：
  * 支援 `Vault (倉庫)` 自動建立。
  * 支援筆記管理（增、刪、改、查），變更毫秒級即時分發至所有線上裝置。
* **🖼️ 附件同步支援**：
  * 完美支援圖片等非筆記檔案同步。
  * 支援大附件分片上傳下載，分片大小可配置，提升同步效率。
* **⚙️ 組態同步**：
  * 支援 `.obsidian` 設定檔的同步。
  * 支援 `PDF` 閱讀進度狀態同步。
* **📝 筆記歷史**：
  * 可以在 Web 頁面、用戶端外掛查看每一個筆記的歷史修改版本。
  * （需服務端 v1.2+ ）
* **🗑️ 資源回收桶**：
  * 支援筆記刪除後，自動進入資源回收桶。
  * 支援從資源回收桶復原筆記。（後續會陸續新增附件復原功能）

* **🚫 離線同步策略**：
  * 支援筆記離線編輯自動合併。（需要外掛端設定）
  * 離線刪除，重新連線之後自動補全或刪除同步。（需要外掛端設定）

* **🔗 分享功能**：
  * 可以 建立/取消 筆記分享。
  * 自動解析分享筆記中參照的圖片、音視訊等附件。
  * 提供分享存取統計功能。
  * 可以設定分享筆記的存取密碼。
  * 可以對分享筆記產生短連結。
* **📂 目錄同步**：
  * 支援資料夾的 建立/重新命名/移動/刪除 同步。

* **🌳 Git 自動化**：
  * 當附件和筆記發生變更時，自動更新並推送到遠端 Git 倉庫。
  * 任務結束後自動釋放系統記憶體。

* **☁️ 多儲存備份與單向鏡像同步**：
  * 適配 S3/OSS/R2/WebDAV/本機 等多種儲存協定。
  * 支援全量/增量 ZIP 定時封存備份。
  * 支援 Vault 資源單向鏡像同步至遠端儲存。
  * 自動清理過期備份，支援自訂保留天數。

* **🗄️ 多資料庫支援**：
  * 原生支援 SQLite、MySQL、PostgreSQL 等多種主流資料庫，滿足從個人到團隊的不同部署需求。

## ☕ 贊助與支援

- 如果覺得此外掛程式很有用，並且想要它繼續開發，請在以下方式支持我：

  | Ko-fi *非中國地區*                                                                               |    | 微信掃碼打賞 *中國地區*                        |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | 或 | <img src="/docs/images/wxds.png" height="150"> |

  - 已支持名單：
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-TW.md">Support.zh-TW.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-TW.md">Support.zh-TW.md (cnb.cool 鏡像庫)</a>

## ⏱️ 更新日誌

- ♨️ [存取查看更新日誌](/docs/CHANGELOG.zh-TW.md)

## 🗺️ 路線圖 (Roadmap)

- [ ] 增加 WebSocket `Protobuf` 傳輸格式的支援，強化同步傳輸效率。
- [ ] 對現有授權機制進行隔離以及優化，提升整體安全性。
- [ ] 增加 Web UI 筆記即時更新。
- [ ] 增加用戶端點對點訊息傳送（非筆記與附件，類似 LocalSend 功能，不支援用戶端儲存，可儲存到服務端）。
- [ ] 各類說明文件完善。
- [ ] 更多的內網穿透（中繼閘道器）的支援。
- [ ] 快速部署計劃：
  * 只需要提供伺服器位址（公網）、帳號密碼即可完成 FNS 服務端的部署。
- [ ] 優化現有的離線筆记合併方案，增加衝突處理機制。

我們正在持續改進，以下是未來的開發計劃：

> **如果您有改進建議或新想法，歡迎透過提交 issue 與我們分享——我們會認真評估並採納合適的建議。**

## 🚀 快速部署

我們提供多種安裝方式，推薦使用 **一鍵指令碼** 或 **Docker**。

### 方式一：一鍵指令碼（推薦）

自動偵測系統環境並完成安裝、服務註冊。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

中國地區可以使用騰訊 `cnb.cool` 鏡像來源：
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**指令碼主要行為：**

  * 自動下載適配目前系統的 Release 二進位檔案。
  * 預設安裝至 `/opt/fast-note`，並在 `/usr/local/bin/fns` 建立全域快捷命令 `fns`。
  * 設定並啟動 Systemd（Linux）或 Launchd（macOS）服務，實作開机自啟。
  * **管理命令**：`fns [install|uninstall|start|stop|status|update|menu]`
  * **互動功能表**：直接執行 `fns` 可進入互動功能表，支援安裝/升級、服務控制、開機自啟設定，以及在 GitHub / CNB 鏡像之間切換。

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
      - "9000:9000"  # RESTful API & WebSocket 連接埠 （其中 /api/user/sync 為 WebSocket 介面位址）
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
    首次存取需註冊帳號。*(如需關閉註冊功能，請在設定檔中設定 `user.register-is-enable: false`)*
3.  **配置用戶端**：
    登入管理面板，點擊 **「複製 API 組態」**。
4.  **連線 Obsidian**：
    開啟 Obsidian 外掛設定頁面，貼上剛才複製的組態資訊即可。


## ⚙️ 組態說明

預設設定檔為 `config.yaml`，程式會自動在 **根目錄** 或 **config/** 目錄下尋找。

查看完整組態範例：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx 反向代理組態範例

查看完整組態範例：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (模型上下文協定) 支援

FNS 現已原生支援 **MCP (Model Context Protocol)**，並同時提供 **SSE** 和 **StreamableHTTP** 兩種傳輸協定。

您可以將 FNS 作為 MCP 服務端直接接入 Cherry Studio、Cursor、Claude Code、hermes-agent 等相容的 AI 用戶端。接入後，AI 即可具備讀寫私人筆記和附件的能力。同時，所有由 MCP 產生的修改，都會透過 WebSocket 即時同步到您的各個裝置終端。

如需使用 Stytch 設定受 OAuth 保護的 MCP 部署，請閱讀 [MCP OAuth 與 Stytch 運行手冊](runbook/mcp-oauth-stytch.zh-TW.md)。

### 通用請求標頭參數

無論使用哪種傳輸模式，均支援以下請求標頭：

- **鑑權 Header**：`Authorization: Bearer <您的 API Token>`（在 WebGUI 的複製 API 組態中取得）
- **選用 Header**：`X-Default-Vault-Name: <筆記庫名稱>`（用於指定 MCP 操作的預設筆記庫，若工具呼叫時未指定 `vault` 參數，則使用此值）
- **選用 Header**：`X-Client: <用戶端類型>`（用於連線 MCP 的用戶端類型，如：`Cherry Studio` / `OpenClaw`）
- **選用 Header**：`X-Client-Version: <用戶端版本>`（用於連線 MCP 的用戶端版本，如：`1.1`）
- **選用 Header**：`X-Client-Name: <用戶端名稱>`（用於連線 MCP 的用戶端名稱，如：`Mac`）

---

### 接入組態：StreamableHTTP 模式（推薦）

StreamableHTTP 是 MCP 生態的標準傳輸協定，單個端點即可完成請求，對防火牆更友善，被較新的 MCP 用戶端（如 Claude Code、hermes-agent）原生支援。

- **介面位址**：`http://<您的伺服器IP或網域>:<連接埠>/api/mcp`
- **請求方式**：`POST`（傳送請求/通知）、`GET`（接聽服務端推送）、`DELETE`（終止會話）

#### 範例：Claude Code / hermes-agent / Cursor 等

*(註：請將 `<ServerIP>`、`<Port>`、`<Token>` 和 `<VaultName>` 替換為您自己的實際資訊)*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp",
      "type": "http",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer <Token>",
        "X-Default-Vault-Name": "<VaultName>",
        "X-Client": "<Client>",
        "X-Client-Version": "<ClientVersion>",
        "X-Client-Name": "<ClientName>"
      }
    }
  }
}
```

---

### 接入組態：SSE 模式（回溯相容）

SSE 模式為舊版傳輸協定，仍完整保留以保持回溯相容，適用於僅支援 SSE 的 MCP 用戶端（如 Cherry Studio）。

- **介面位址**：`http://<您的伺服器IP或網域>:<連接埠>/api/mcp/sse`

#### 範例：Cherry Studio / Cline 等

*(註：請將 `<ServerIP>`、`<Port>`、`<Token>` 和 `<VaultName>` 替換為您自己的實際資訊)*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp/sse",
      "type": "sse",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer <Token>",
        "X-Default-Vault-Name": "<VaultName>",
        "X-Client": "<Client>",
        "X-Client-Version": "<ClientVersion>",
        "X-Client-Name": "<ClientName>"
      }
    }
  }
}
```

## 🔗 用戶端 & 用戶端外掛程式 & 協作專案

* Obsidian Fast Note Sync 外掛程式
  * [Obsidian Fast Note Sync 外掛程式](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool 鏡像庫](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* 第三方用戶端
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) 一款基於 Python 和 FNS WebSocket 同步協定的雙向即時同步命令列用戶端，適用於無 GUI 的 Linux 伺服器環境（如 OpenClaw），實現與 Obsidian 桌面/行動端等價的同步能力。
  * [go-fast-note-sync](https://github.com/erichll/go-fast-note-sync) 一款基於 Go 和 FNS WebSocket 同步協定的 Go CLI 背景同步精靈程序，主要面向 Linux 無頭 (headless) 環境，同時也支援 macOS 和 Windows。
  * [Fast-note-sync-docker](https://github.com/youpingfang/obsidian-note-sync-docker) 一款基於 Docker、Python、FNS WebSocket 同步協定的快速容器化部署方案，實現筆記庫和設定檔到遠端伺服器。
* 協作專案
  * [Share to Save](https://github.com/chenxiccc/Obsidian-Share-to-Save) 是一款 Obsidian 外掛程式，將分享的網頁 URL 自動下載為 Markdown 筆記。
