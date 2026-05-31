[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

何かご不明な点や問題がございましたら、[issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) を作成するか、Telegramコミュニティグループにてご相談ください: [https://t.me/obsidian_users](https://t.me/obsidian_users)

中国本土地域では、Tencent `cnb.cool` ミラーリポジトリの使用をお勧めします: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


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
  <em>Golang + WebSocket + React で構築</em>
</p>

<p align="center">
  クライアントデータ提供には、プラグインとの併用が必要です：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

* **🧰 MCP (Model Context Protocol) ネイティブサポート**：
  * `FNS` は MCP サーバーとして `Cherry Studio` や `Cursor` などの互換性のある AI クライアントに接続できます。これにより、AI が個人のノートや添付ファイルを読み書きできるようになり、すべての変更が各デバイスへリアルタイムに同期されます。
* **🚀 REST API サポート**：
  * 標準の REST API インターフェースを提供し、自動化スクリプトや AI アシスタント統合などのプログラムから Obsidian ノートの作成・読み取り・更新・削除（CRUD）操作をサポートします。
  * 詳細は [RESTful API ドキュメント](/docs/REST_API.md) または [OpenAPI ドキュメント](/docs/swagger.yaml) をご参照ください。
* **💻 Web 管理パネル**：
  * モダンな管理画面を内蔵しており、ユーザーの作成、プラグイン設定の生成、リポジトリやノートコンテンツの管理が簡単に行えます。
* **🔄 マルチデバイスノート同期**：
  * **Vault (保管庫)** の自動作成をサポート。
  * ノート管理（追加、削除、変更、検索）をサポートし、変更はミリ秒単位でリアルタイムにすべてのオンラインデバイスへ配信されます。
* **🖼️ 添付ファイル同期サポート**：
  * 画像などのノート以外のファイルの同期を完全にサポート。
  * 大容量の添付ファイルの分割アップロード・ダウンロードをサポートし、分割サイズの設定も可能です。これにより同期の効率が向上します。
* **⚙️ 設定同期**：
  * `.obsidian` 設定ファイルの同期をサポート。
  * `PDF` 閲覧の進行状況の同期をサポート。
* **📝 ノート履歴**：
  * Webページやプラグイン端から、各ノートの過去の変更履歴（バージョン）を確認できます。（サーバー v1.2+ が必要です）
* **🗑️ ゴミ箱**：
  * ノート削除後、自動的にゴミ箱へ移動します。
  * ゴミ箱からのノート復元をサポート（添付ファイルの復元機能も順次追加予定です）。

* **🚫 オフライン同期ポリシー**：
  * オフライン編集ノートの自動マージをサポート（プラグイン側での設定が必要です）。
  * オフライン削除について、再接続後の自動補完または削除同期をサポート（プラグイン側での設定が必要です）。

* **🔗 共有機能**：
  * ノートの共有設定の作成・解除が可能です。
  * 共有されたノート内で参照されている画像、音声、動画などの添付ファイルを自動的に解析します。
  * 共有ノートのアクセス統計機能を提供。
  * 共有ノートにアクセスパスワードを設定できます。
  * 共有ノートの短縮URLを生成できます。
* **📂 ディレクトリ同期**：
  * フォルダの作成、名前変更、移動、削除の同期をサポート。

* **🌳 Git 自動化**：
  * 添付ファイルやノートに変更があった際、自動的に更新してリモートの Git リポジトリへプッシュします。
  * タスク終了後、自動的にシステムメモリを解放します。

* **☁️ 複数ストレージへのバックアップと一方向ミラー同期**：
  * S3、OSS、R2、WebDAV、ローカルなどの多様なストレージプロトコルに対応。
  * 定期的なZIP形式でのフル／差分アーカイブバックアップをサポート。
  * Vault リソースのリモートストレージへの一方向ミラー同期をサポート。
  * 期限切れバックアップの自動クリーンアップ（保存日数のカスタマイズ可能）をサポート。

* **🗄️ 複数データベースのサポート**：
  * SQLite、MySQL、PostgreSQL などの主要なデータベースをネイティブサポートし、個人からチームまでの多様なデプロイニーズに対応します。

## ☕ スポンサーとサポート

- もしこのプラグインが役立ち、今後の開発を支援したいと思われる場合は、以下の方法でご支援をお願いいたします：

  | Ko-fi *中国本土以外*                                                                               |    | 微信（WeChat）スキャン決済 *中国本土*          |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | または | <img src="/docs/images/wxds.png" height="150"> |

  - 支援者リスト：
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md (cnb.cool ミラーリポジトリ)</a>

## ⏱️ 更新履歴

- ♨️ [更新履歴を表示する](/docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ (Roadmap)

- [ ] WebSocket `Protobuf` 転送フォーマットのサポートを追加し、同期転送の効率を強化。
- [ ] 既存の認証メカニズムを隔離・最適化し、全体的な安全性を向上。
- [ ] WebGui ノートのリアルタイム更新を追加。
- [ ] クライアント間のピアツーピア（P2P）メッセージ送信を追加（ノート・添付ファイル以外、LocalSendに類似する機能。クライアント側での保存は非サポート、サーバー側への保存は可能）。
- [ ] 各種ヘルプドキュメントの充実。
- [ ] より多くのイントラネット浸透（中継ゲートウェイ）のサポート。
- [ ] クイックデプロイプラン
  * サーバーアドレス（パブリックIP）、アカウント、パスワードを提供するだけで FNS サーバーのデプロイが完了します。
- [ ] 既存のオフラインノートマージスキームを最適化し、競合処理メカニズムを追加。

私たちは継続的に改善を行っています。以下は今後の開発計画です：

> **改善の提案や新しいアイデアがございましたら、issue を通じてお気軽にご共有ください。内容を精査し、適切なアイデアを採用させていただきます。**

## 🚀 クイックデプロイ

複数のインストール方法を提供しています。**ワンクリックインストールスクリプト** または **Docker** の使用をお勧めします。

### 方法一：ワンクリックスクリプト（推奨）

システム環境を自動検出してインストールし、サービスとして登録します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

中国本土地域では、Tencent `cnb.cool` ミラーリポジトリを使用できます。
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**スクリプトの主な動作：**

  * 現在のシステムに適した Release バイナリファイルを自動的にダウンロードします。
  * デフォルトで `/opt/fast-note` にインストールされ、`/usr/local/bin/fns` にグローバルショートカットコマンド `fns` を作成します。
  * Systemd（Linux）または Launchd（macOS）サービスを設定・起動し、自動起動を実現します。
  * **管理コマンド**：`fns [install|uninstall|start|stop|status|update|menu]`
  * **対話型メニュー**：`fns` を直接実行すると対話型メニューに入り、インストール／アップグレード、サービス制御、自動起動の設定、および GitHub / CNB ミラー間の切り替えが可能です。

-----

### 方法二：Docker デプロイ

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

`docker-compose.yaml` ファイルを作成します：

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # RESTful API & WebSocket ポート。/api/user/sync が WebSocket インターフェースアドレスです。
    volumes:
      - ./storage:/fast-note-sync/storage  # データストレージ
      - ./config:/fast-note-sync/config    # 設定ファイル
```

サービスの起動：

```bash
docker compose up -d
```

-----

### 方法三：手動バイナリインストール

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) から対応するシステムの最新バージョンをダウンロードし、解凍後に以下を実行します：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用ガイド

1.  **管理パネルへのアクセス**：
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**：
    初回アクセス時はアカウントの登録が必要です。*(一般登録機能を無効にする場合は、設定ファイルで `user.register-is-enable: false` に設定してください)*
3.  **クライアントの設定**：
    管理パネルにログインし、**「API設定をコピー」**をクリックします。
4.  **Obsidianとの接続**：
    Obsidianのプラグイン設定ページを開き、コピーした設定情報を貼り付けます。


## ⚙️ 設定について

デフォルトの設定ファイルは `config.yaml` です。プログラムは **ルートディレクトリ** または **config/** ディレクトリから自動的に検索します。

設定例：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx リバースプロキシ設定例

設定例：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) サポート

FNS は **MCP (Model Context Protocol)** をネイティブサポートしており、**SSE** と **StreamableHTTP** の2つの転送プロトコルを同時に提供します。

FNS を MCP サーバーとして Cherry Studio、Cursor、Claude Code、hermes-agent などの互換性のある AI クライアントに直接接続できます。接続後、AI は個人のノートや添付ファイルを読み書きする能力を持ちます。同時に、MCP によって生成されたすべての変更は、WebSocket を介して各デバイス端末にリアルタイムで同期されます。

### 共通リクエストヘッダーパラメータ

どの転送モードを使用する場合でも、以下のリクエストヘッダーをサポートします：

- **認証ヘッダー**：`Authorization: Bearer <APIトークン>`（WebGUIの「API設定のコピー」から取得可能）
- **オプションヘッダー**：`X-Default-Vault-Name: <保管庫名>`（MCP操作のデフォルト保管庫を指定。ツール呼び出し時に `vault` パラメータが指定されていない場合、この値が使用されます）
- **オプションヘッダー**：`X-Client: <クライアントタイプ>`（MCPに接続するクライアントのタイプ。例：Cherry Studio / OpenClaw）
- **オプションヘッダー**：`X-Client-Version: <クライアントバージョン>`（MCPに接続するクライアントのバージョン。例：1.1）
- **オプションヘッダー**：`X-Client-Name: <クライアント名>`（MCPに接続するクライアントのデバイス名。例：Mac）

---

### 接続設定：StreamableHTTP モード（推奨）

StreamableHTTP は MCP エコシステムの標準転送プロトコルです。単一のエンドポイントでリクエストを完了できるため、ファイアウォールに対して優しく、新しい MCP クライアント（Claude Code、hermes-agent など）でネイティブにサポートされています。

- **インターフェースアドレス**：`http://<サーバーIPまたはドメイン>:<ポート>/api/mcp`
- **リクエストメソッド**：`POST`（リクエスト/通知の送信）、`GET`（サーバーからのプッシュ監視）、`DELETE`（セッション終了）

#### 例：Claude Code / hermes-agent / Cursor など

*(注： `<ServerIP>`、`<Port>`、`<Token>`、および `<VaultName>` をご自身の実際の情報に置き換えてください)*

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

SSE モードは従来の転送プロトコルですが、後方互換性を維持するために完全に保留されており、SSE のみをサポートする MCP クライアント（Cherry Studio など）に適しています。

- **インターフェースアドレス**：`http://<サーバーIPまたはドメイン>:<ポート>/api/mcp/sse`

#### 例：Cherry Studio / Cline など

*(注： `<ServerIP>`、`<Port>`、`<Token>`、および `<VaultName>` をご自身の実際の情報に置き換えてください)*

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

## 🔗 クライアント ＆ クライアントプラグイン

* Obsidian Fast Note Sync プラグイン
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool ミラーリポジトリ](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* サードパーティクライアント
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) Python と FNS WebSocket 同期プロトコルに基づいた双方向リアルタイム同期コマンドラインクライアント。GUIのない Linux サーバー環境（OpenClawなど）に最適で、Obsidian デスクトップ/モバイル端と同等の同期能力を実現します。
  * [go-fast-note-sync](https://github.com/erichll/go-fast-note-sync) Go と FNS WebSocket 同期プロトコルに基づいた Go CLI バックグラウンド同期デーモンプロセス。主にヘッドレス（headless）な Linux 環境向けですが、macOS と Windows もサポートしています。
  * [Fast-note-sync-docker](https://github.com/youpingfang/obsidian-note-sync-docker) Docker、Python、FNS WebSocket 同期プロトコルに基づいた迅速なコンテナ化デプロイソリューション。ノート保管庫と設定ファイルをリモートサーバーへ同期します。
