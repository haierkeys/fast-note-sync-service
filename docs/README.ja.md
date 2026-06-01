[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

質問がある場合は [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) を作成するか、テレグラムグループに参加してヘルプを求めてください: [https://t.me/obsidian_users](https://t.me/obsidian_users)

中国本土のユーザーは、Tencent cnb.cool ミラーリポジトリの使用をお勧めします: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能・低遅延のノート同期、オンライン管理、リモート REST API サービスプラットフォーム</strong>
  <br>
  <em>Golang + Websocket + React ベース</em>
</p>

<p align="center">
  データ提供はクライアントプラグインと組み合わせる必要があります：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## 🎯 主な機能

* **🧰 MCP (Model Context Protocol) ネイティブサポート**:
  * `FNS` はMCPサーバーとして、`Cherry Studio`や`Cursor`などの互換性のあるAIクライアントに接続でき、AIにプライベートなノートや添付ファイルの読み書き機能を提供します。すべての変更は各クライアントにリアルタイムで同期されます。
* **🚀 REST API サポート**:
  * 标准のREST APIインターフェースを提供し、プログラム（自動化スクリプトやAIアシスタントの統合など）を介したObsidianノートのCRUD操作をサポートします。
  * 詳細は、[RESTful API ドキュメント](/docs/REST_API.md) または [OpenAPI ドキュメント](/docs/swagger.yaml) を参照してください。
* **💻 Web管理パネル**:
  * 最新の管理画面を内蔵し、ユーザーの作成、プラグイン設定の生成、リポジトリやノート内容の管理を簡単に行えます。
* **🔄 複数デバイスのノート同期**:
  * Vault（リポジトリ）の自動作成をサポート。
  * ノート管理（追加、削除、編集、検索）をサポートし、変更はミリ秒単位でリアルタイムにすべてのオンラインデバイスに配信されます。
* **🖼️ 添付ファイル同期サポート**:
  * 画像などのノート以外のファイルの同期を完全にサポート。
  * 大容量の添付ファイルの分割アップロード・ダウンロードをサポートし、分割サイズを設定して同期効率を向上できます。
* **⚙️ 設定同期**:
  * `.obsidian` 設定ファイルの同期をサポート。
  * `PDF` の閲覧進捗ステータスの同期をサポート。
* **📝 ノート履歴**:
  * Webページやクライアントプラグイン側で、各ノートの修正履歴バージョンを確認できます。
  * （サーバー側 v1.2+ が必要）
* **🗑️ ゴミ箱**:
  * ノート削除後、自動的にゴミ箱に送られます。
  * ゴミ箱からのノート復元をサポート。（添付ファイルの復元機能は今後順次追加予定です）

* **🚫 オフライン同期ポリシー**:
  * ノートのオフライン編集時の自動マージをサポート。（プラグイン側の設定が必要です）
  * オフラインでの削除後、再接続時に自動的に補完または同期削除されます。（プラグイン側の設定が必要です）

* **🔗 共有機能**:
  * ノート共有の作成/キャンセルが可能。
  * 共有ノート内で引用されている画像、音声、動画などの添付ファイルを自動的に解析します。
  * 共有アクセス統計機能を提供。
  * 共有ノートのアクセスパスワードを設定可能。
  * 共有ノートの短縮リンクを生成可能。
* **📂 フォルダ同期**:
  * フォルダの作成/名前変更/移動/削除の同期をサポート。

* **🌳 Git自動化**:
  * 添付ファイルやノートに変更があった場合、自動的に更新してリモートのGitリポジトリにプッシュします。
  * タスク終了後、自動的にシステムメモリを解放します。

* **☁️ マルチストレージバックアップと一方向ミラー同期**:
  * S3/OSS/R2/WebDAV/ローカルなどの複数のストレージプロトコルに対応。
  * フル/増分ZIP定期アーカイブバックアップをサポート。
  * Vaultリソースのリモートストレージへの一方向ミラー同期をサポート。
  * 期限切れバックアップの自動クリーンアップをサポートし、カスタム保持日数を設定可能。

* **🗄️ マルチデータベースサポート**:
  * SQLite、MySQL、PostgreSQLなどの主要データベースをネイティブサポートし、個人からチームまでのさまざまなデプロイ要件に対応します。

## ☕ スポンサーとサポート

- このプラグインが便利で、開発の継続をサポートしたい場合は、以下の方法で支援をお願いします:

  | Ko-fi *中国以外の地域*                                                                          |    | WeChat Pay *中国国内*                          |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - 支援者リスト：
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.ja.md">Support.ja.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.ja.md">Support.ja.md (cnb.cool ミラーリポジトリ)</a>

## ⏱️ 更新履歴

- ♨️ [更新履歴を表示する](/docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ

- [ ] WebSocketの `Protobuf` 転送フォーマットのサポートを追加し、同期伝送効率を強化。
- [ ] 既存の認可メカニズムを分離および最適化し、全体のセキュリティを向上。
- [ ] Web UIにおけるノートのリアルタイム更新を追加。
- [ ] クライアント間のピアツーピア（P2P）メッセージ送信を追加（ノートや添付ファイル以外、LocalSendのような機能。クライアント側での保存はサポートせず、サーバー側での保存が可能）。
- [ ] 各種ヘルプドキュメントの充実。
- [ ] より多くのイントラネット浸透（リレーゲートウェイ）のサポート。
- [ ] 迅速なデプロイ計画：
  * サーバーアドレス（パブリックIP）とアカウント名・パスワードを提供するだけで、FNSサーバー側のデプロイが完了します。
- [ ] 既存のオフラインノートマージスキームを最適化し、競合処理メカニズムを追加。

私たちは継続的な改善を行っており、以下は今後の開発計画です：

> **改善の提案や新しいアイデアがございましたら、issueを提出して共有してください。適切な提案は慎重に評価し採用いたします。**

## 🚀 クイックデプロイ

複数のインストール方法を提供しています。**ワンクリックスクリプト** または **Docker** の使用を推奨します。

### 方法1：ワンクリックスクリプト（推奨）

システム環境を自动検出し、インストールとサービス登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

中国のユーザーは、Tencent `cnb.cool` ミラーソースを使用できます：
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**スクリプトの主な動作：**

  * 現在のシステムに適したReleaseバイナリファイルを自動的にダウンロードします。
  * デフォルトで `/opt/fast-note` にインストールされ、`/usr/local/bin/fns` にグローバルなショートカットコマンド `fns` を作成します。
  * Systemd（Linux）または Launchd（macOS）サービスを設定・起動し、自動起動を有効にします。
  * **管理コマンド**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **インタラクティブメニュー**: 直接 `fns` を実行するとメニューに入り、インストール/アップグレード、サービス制御、自動起動設定、GitHubとCNBミラー間の切り替えなどをサポートします。

-----

### 方法2：Dockerデプロイ

#### Docker Run

```bash
# 1. イメージのプル
docker pull haierkeys/fast-note-sync-service:latest

# 2. コンテナの起動
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

docker-compose.yaml ファイルを作成：

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # RESTful API & WebSocket ポート（うち /api/user/sync が WebSocket インターフェースアドレス）
    volumes:
      - ./storage:/fast-note-sync/storage  # 数据ストレージ
      - ./config:/fast-note-sync/config    # 設定ファイル
```

サービスの起動：

```bash
docker compose up -d
```

-----

### 方法3：手動バイナリインストール

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) から対応する系统的最新バージョンをダウンロードし、解凍して実行します：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用ガイド

1.  **管理パネルにアクセス**:
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**:
    初回アクセス時はアカウント登録が必要です。（登録機能を無効にする場合は、設定ファイルで `user.register-is-enable: false` と設定してください）
3.  **クライアントの設定**:
    管理パネルにログインし、「API設定をコピー」をクリックします。
4.  **Obsidianへの接続**:
    Obsidianのプラグイン設定ページを開き、コピーした設定情報を貼り付けます。


## ⚙️ 設定の説明

デフォルトの設定ファイルは `config.yaml` です。プログラムは自動的に**ルートディレクトリ**または **config/** ディレクトリ内を検索します。

完全な設定例を表示：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx リバースプロキシ設定例

完全な設定例を表示：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) サポート

FNSは現在、**MCP (Model Context Protocol)** をネイティブサポートしており、**SSE** と **StreamableHTTP** の両方の転送プロトコルを提供しています。

FNSをMCPサーバーとしてCherry Studio、Cursor、Claude Code、hermes-agentなどの互換性のあるAIクライアントに直接接続できます。接続後、AIはプライベートなノートや添付ファイルを読み書きできるようになります。また、MCPによる変更はWebSocketを通じてリアルタイムに各デバイス端末に同期されます。

### 共通リクエストヘッダーパラメータ

どの転送モードを使用する場合でも、以下のリクエストヘッダーがサポートされます：

- **認証用Header**: `Authorization: Bearer <APIトークン>` （WebGUIの「API設定をコピー」から取得）
- **オプションのHeader**: `X-Default-Vault-Name: <ノート庫名>` （MCP操作のデフォルトのノート庫を指定。ツール呼び出し時に `vault` パラメータが指定されていない場合、この値が使用されます）
- **オプションのHeader**: `X-Client: <クライアントタイプ>` （MCPに接続するクライアントのタイプ。例：`Cherry Studio`、`OpenClaw`）
- **オプションのHeader**: `X-Client-Version: <クライアントバージョン>` （接続するクライアントのバージョン。例：`1.1`）
- **オプションのHeader**: `X-Client-Name: <クライアント名>` （接続するクライアントの名前。例：`Mac`）

---

### 接続設定: StreamableHTTP モード（推奨）

StreamableHTTPはMCPエコシステムの標準的な転送プロトコルです。単一のエンドポイントでリクエストを完了でき、ファイアウォールに優しく、新しいMCPクライアント（Claude Codeやhermes-agentなど）でネイティブにサポートされています。

- **インターフェースアドレス**: `http://<サーバーIPまたはドメイン>:<ポート>/api/mcp`
- **リクエスト方法**: `POST`（リクエスト/通知送信）、`GET`（サーバーからのプッシュ監視）、`DELETE`（セッション終了）

#### 例：Claude Code / hermes-agent / Cursor など

*(注: `<ServerIP>`、`<Port>`、`<Token>`、`<VaultName>` をご自身の実際の情報に置き換えてください)*

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

### 接続設定: SSE モード（後方互換）

SSEモードは従来の転送プロトコルです。後方互換性を維持するために完全に保持されており、SSEのみをサポートするMCPクライアント（Cherry Studioなど）に適しています。

- **インターフェースアドレス**: `http://<サーバーIPまたはドメイン>:<ポート>/api/mcp/sse`

#### 例：Cherry Studio / Cline など

*(注: `<ServerIP>`、`<Port>`、`<Token>`、`<VaultName>` をご自身の実際の情報に置き換えてください)*

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

## 🔗 クライアント ＆ クライアントプラグイン ＆ 共同プロジェクト

* Obsidian Fast Note Sync プラグイン
  * [Obsidian Fast Note Sync プラグイン](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool ミラーリポジトリ](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* サードパーティ製クライアント
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) PythonおよびFNS WebSocket同期プロトコルに基づき実装された、双方向リアルタイム同期のコマンドラインクライアント。GUIのないLinuxサーバー環境（OpenClawなど）に適しており、Obsidianデスクトップ/モバイル端と同等の同期機能を実現します。
  * [go-fast-note-sync](https://github.com/erichll/go-fast-note-sync) GoおよびFNS WebSocket同期プロトコルに基づいたGo CLIバックグラウンド同期デーモン。主にLinuxヘッドレス環境向けであり、macOSとWindowsもサポートしています。
  * [Fast-note-sync-docker](https://github.com/youpingfang/obsidian-note-sync-docker) Docker、Python、FNS WebSocket同期プロトコルに基づいた迅速なコンテナ化デプロイソリューション。ノートライブラリと設定ファイルをリモートサーバーに同期します。
* 共同プロジェクト
  * [Share to Save](https://github.com/chenxiccc/Obsidian-Share-to-Save) 共有されたウェブページのURLを自動的にMarkdownノートとしてダウンロードするObsidianプラグイン。
