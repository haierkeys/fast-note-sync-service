[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

問題が発生した場合は、新しい [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) を作成するか、Telegramグループに参加してサポートを求めてください: [https://t.me/obsidian_users](https://t.me/obsidian_users)

中国本土のユーザーには、Tencent `cnb.cool` ミラーの使用を推奨します: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高パフォーマンス・低レイテンシのノート同期、オンライン管理、リモート REST API サービスプラットフォーム</strong>
  <br>
  <em>Golang + WebSocket + React で構築</em>
</p>

<p align="center">
  データ同期にはクライアントプラグインが必要です：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## 🎯 主要機能

* **🧰 MCP (Model Context Protocol) ネイティブサポート**：
  * `FNS` は MCP サーバーとして `Cherry Studio`、`Cursor` などの対応 AI クライアントに接続でき、AI がプライベートノートや添付ファイルを読み書きする能力を持ち、すべての変更はリアルタイムで各デバイスに同期されます。
* **🚀 REST API サポート**：
  * 標準 REST API エンドポイントを提供し、プログラムによるアクセス（自動化スクリプト、AI アシスタント統合など）で Obsidian ノートの CRUD 操作をサポートします。
  * 詳細は [RESTful API ドキュメント](/docs/REST_API.md) または [OpenAPI ドキュメント](/docs/swagger.yaml) をご参照ください。
* **💻 Web 管理パネル**：
  * モダンな管理インターフェースを内蔵し、ユーザー作成、プラグイン設定の生成、Vault やノートコンテンツの管理を簡単に行えます。
* **🔄 マルチデバイスノート同期**：
  * **Vault（ノートリポジトリ）** の自動作成をサポート。
  * ノート管理（作成・削除・更新・検索）をサポートし、ミリ秒単位でオンラインの全デバイスにリアルタイム配信します。
* **🖼️ 添付ファイル同期サポート**：
  * 画像などのノート以外のファイルの同期を完全サポート。
  * 大容量添付ファイルのチャンク分割アップロード/ダウンロードをサポートし、チャンクサイズは設定可能で同期効率を向上させます。
* **⚙️ 設定同期**：
  * `.obsidian` 設定ファイルの同期をサポート。
  * `PDF` の閲覧進捗状態の同期をサポート。
* **📝 ノート履歴**：
  * Web パネルおよびプラグイン側で各ノートの過去の修正バージョンを確認できます。
  * (サーバー v1.2+ が必要)
* **🗑️ ゴミ箱**：
  * 削除されたノートは自動的にゴミ箱に移動します。
  * ゴミ箱からのノート復元をサポート。(添付ファイルの復元機能は今後追加予定)

* **🚫 オフライン同期戦略**：
  * オフライン編集したノートの自動マージをサポート。(プラグイン側の設定が必要)
  * オフライン中の削除は再接続後に自動的に補完または削除同期されます。(プラグイン側の設定が必要)

* **🔗 共有機能**：
  * ノート共有リンクの作成/取り消しが可能。
  * 共有ノート内で参照されている画像、音声、動画などの添付ファイルを自動解析。
  * 共有アクセス統計機能を提供。
  * 共有ノートへのアクセスパスワードの設定をサポート。
  * 共有ノートの短縮リンク生成をサポート。
* **📂 ディレクトリ同期**：
  * フォルダの作成/名前変更/移動/削除の同期をサポート。

* **🌳 Git 自動化**：
  * 添付ファイルやノートが変更されると、自動的にリモート Git リポジトリにコミット・プッシュします。
  * タスク完了後、システムメモリを自動解放します。

* **☁️ マルチストレージバックアップ & 一方向ミラー同期**：
  * S3、OSS、R2、WebDAV、ローカルなど複数のストレージプロトコルに対応。
  * フル/増分 ZIP アーカイブの定期バックアップをサポート。
  * Vault リソースのリモートストレージへの一方向ミラー同期をサポート。
  * 期限切れバックアップの自動クリーンアップ、保持日数のカスタマイズをサポート。

* **🗄️ マルチデータベースサポート**：
  * SQLite、MySQL、PostgreSQL などの主流データベースをネイティブサポートし、個人からチームまで多様なデプロイニーズに対応。

## ☕ スポンサーとサポート

- このプロジェクトが役に立ち、継続的な開発を支援したい場合は、以下の方法でサポートをご検討ください：

  | Ko-fi *（中国本土以外）*                                                                           |     | WeChat 寄付 *（中国本土）*                       |
  |--------------------------------------------------------------------------------------------------|-----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | または | <img src="/docs/images/wxds.png" height="150"> |

  - サポーター一覧：
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md (cnb.cool ミラー)</a>

## ⏱️ 更新履歴

- ♨️ [更新履歴を見る](/docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ

- [ ] WebSocket `Protobuf` 転送フォーマットのサポートを追加し、同期転送効率を強化。
- [ ] 既存の認証メカニズムを分離・最適化し、全体的なセキュリティを向上。
- [ ] WebGui にノートリアルタイム更新機能を追加。
- [ ] クライアント間のピアツーピアメッセージ送信を追加（ノート・添付ファイル以外、localsend 類似機能、クライアント保存不可、サーバー保存可）。
- [ ] 各種ヘルプドキュメントの充実。
- [ ] より多くのイントラネット貫通（リレーゲートウェイ）のサポート。
- [ ] クイックデプロイ計画：
  * サーバーアドレス（公開）、アカウント、パスワードを提供するだけで FNS サーバーのデプロイが完了。
- [ ] 既存のオフラインノートマージ方式を最適化し、コンフリクト解決メカニズムを追加。

継続的に改善しています。以下は今後の開発計画です：

> **改善提案や新しいアイデアがある場合は、issue を提出して共有してください——適切な提案を真剣に評価・採用します。**

## 🚀 クイックデプロイ

複数のインストール方法を提供しています。**ワンクリックスクリプト** または **Docker** を推奨します。

### 方法 1：ワンクリックスクリプト（推奨）

システム環境を自動検出し、インストールとサービス登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

中国本土のユーザーは Tencent `cnb.cool` ミラーを使用できます：
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**スクリプトの主な動作：**

  * 現在のシステムに対応した Release バイナリを自動ダウンロード。
  * デフォルトで `/opt/fast-note` にインストールし、`/usr/local/bin/fns` にグローバルショートカットコマンド `fns` を作成。
  * Systemd（Linux）または Launchd（macOS）サービスを設定・起動し、起動時自動起動を実現。
  * **管理コマンド**：`fns [install|uninstall|start|stop|status|update|menu]`
  * **インタラクティブメニュー**：`fns` を直接実行するとインタラクティブメニューに入り、インストール/アップグレード、サービス制御、起動設定、GitHub / CNB ミラーの切り替えをサポート。

-----

### 方法 2：Docker デプロイ

#### Docker Run

```bash
# 1. イメージをプル
docker pull haierkeys/fast-note-sync-service:latest

# 2. コンテナを起動
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

`docker-compose.yaml` ファイルを作成：

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # RESTful API & WebSocket ポート；/api/user/sync が WebSocket エンドポイント
    volumes:
      - ./storage:/fast-note-sync/storage  # データストレージ
      - ./config:/fast-note-sync/config    # 設定ファイル
```

サービスを起動：

```bash
docker compose up -d
```

-----

### 方法 3：手動バイナリインストール

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) から対応 OS の最新バージョンをダウンロードし、解凍後に実行：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用ガイド

1.  **管理パネルへのアクセス**：
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**：
    初回アクセス時にアカウントを登録します。*(登録機能を無効にするには、設定ファイルで `user.register-is-enable: false` を設定してください)*
3.  **クライアント設定**：
    管理パネルにログインし、**「API 設定をコピー」** をクリックします。
4.  **Obsidian に接続**：
    Obsidian プラグイン設定ページを開き、コピーした設定情報を貼り付けます。


## ⚙️ 設定

デフォルトの設定ファイルは `config.yaml` で、プログラムは **ルートディレクトリ** または **config/** ディレクトリを自動検索します。

完全な設定例を確認：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx リバースプロキシ設定例

完全な設定例を確認：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) サポート

FNS は **MCP (Model Context Protocol)** をネイティブサポートし、**SSE** と **StreamableHTTP** の両方の転送プロトコルを提供します。

FNS を MCP サーバーとして Cherry Studio、Cursor、Claude Code、hermes-agent などの対応 AI クライアントに直接接続できます。接続後、AI はプライベートノートと添付ファイルを読み書きする能力を持ちます。MCP によって生成されたすべての変更は、WebSocket を介してリアルタイムで各デバイスに同期されます。

### 共通リクエストヘッダー

転送モードにかかわらず、以下のヘッダーがサポートされています：

- **認証ヘッダー**：`Authorization: Bearer <API Token>` （WebGUI の API 設定コピーから取得）
- **オプションヘッダー**：`X-Default-Vault-Name: <Vault名>` （MCP 操作のデフォルト Vault を指定；ツール呼び出し時に `vault` パラメータが指定されていない場合に使用）
- **オプションヘッダー**：`X-Client: <クライアントタイプ>` （MCP 接続のクライアントタイプ、例：Cherry Studio / OpenClaw）
- **オプションヘッダー**：`X-Client-Version: <クライアントバージョン>` （MCP 接続のクライアントバージョン、例：1.1）
- **オプションヘッダー**：`X-Client-Name: <クライアント名>` （MCP 接続のクライアント名、例：Mac）

---

### 接続設定：StreamableHTTP モード（推奨）

StreamableHTTP は MCP エコシステムの標準転送プロトコルです。単一エンドポイントですべてのリクエストを処理し、ファイアウォールに対してよりフレンドリーで、新しい MCP クライアント（Claude Code、hermes-agent など）にネイティブサポートされています。

- **エンドポイント**：`http://<サーバーIPまたはドメイン>:<ポート>/api/mcp`
- **メソッド**：`POST`（リクエスト/通知送信）、`GET`（サーバー送信イベント待機）、`DELETE`（セッション終了）

#### 例：Claude Code / hermes-agent / Cursor など

*（注：`<ServerIP>`、`<Port>`、`<Token>`、`<VaultName>` をご自身の実際の情報に置き換えてください）*

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

### 接続設定：SSE モード（後方互換）

SSE モードはレガシー転送プロトコルで、後方互換性のために完全に保持されています。SSE のみをサポートする MCP クライアント（Cherry Studio など）に適しています。

- **エンドポイント**：`http://<サーバーIPまたはドメイン>:<ポート>/api/mcp/sse`

#### 例：Cherry Studio / Cline など

*（注：`<ServerIP>`、`<Port>`、`<Token>`、`<VaultName>` をご自身の実際の情報に置き換えてください）*

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

## 🔗 クライアント & クライアントプラグイン

* Obsidian Fast Note Sync プラグイン
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool ミラー](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* サードパーティクライアント
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) — Python ベースで FNS WebSocket API を介した双方向リアルタイム同期を実装したコマンドラインクライアント。GUI のない Linux サーバー環境（OpenClaw など）向けに設計され、Obsidian デスクトップ/モバイルクライアントと同等の同期能力を提供します。
