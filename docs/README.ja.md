[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

ご質問がある場合は、新しい [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) を作成するか、Telegramの交流グループに参加して助けを求めてください: [https://t.me/obsidian_users](https://t.me/obsidian_users)

中国本土地域では、Tencentの `cnb.cool` ミラーリポジトリの使用を推奨します: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能・低遅延なノート同期、オンライン管理、リモートREST APIサービスプラットフォーム</strong>
  <br>
  <em>Golang + Websocket + Reactで構築</em>
</p>

<p align="center">
  データを利用するには、クライアントプラグインを併用する必要があります：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## 🎯 コア機能

* **🧰 MCP (Model Context Protocol) のネイティブサポート**：
  * `FNS` はMCPサーバーとして `Cherry Studio` や `Cursor` などの互換性のあるAIクライアントに接続できます。これにより、AIはあなたのプライベートノートや添付ファイルの読み書き能力を持ち、すべての変更が即座に各デバイスに同期されます。
* **🚀 REST APIのサポート**：
  * 標準的なREST APIインターフェースを提供し、プログラム（自動化スクリプトやAIアシスタントの統合など）によるObsidianノートの作成、読み取り、更新、削除をサポートします。
  * 詳細は [RESTful API ドキュメント](/docs/REST_API.md) または [OpenAPI ドキュメント](/docs/swagger.yaml) を参照してください。
* **💻 Web管理パネル**：
  * モダンな管理インターフェースを内蔵し、ユーザーの作成、プラグイン設定の生成、Vaultやノート内容の管理を簡単に行うことができます。
* **🔄 マルチデバイスによるノート同期**：
  * **Vault (保管庫)** の自動作成をサポート。
  * ノート管理（追加、削除、変更、検索）をサポートし、変更内容をミリ秒レベルでオンラインの全デバイスにリアルタイム配信します。
* **🖼️ 添付ファイル同期のサポート**：
  * 画像などの非ノートファイルの同期を完全にサポート。
  * 大規模な添付ファイルの分割アップロード・ダウンロードをサポートし、分割サイズの設定も可能で、同期効率を向上させます。
* **⚙️ 設定の同期**：
  * `.obsidian` 設定ファイルの同期をサポート。
  * `PDF` の進行状況の同期をサポート。
* **📝 ノートの履歴機能**：
  * Webページの管理画面およびプラグイン側から、各ノートの歴史的な変更バージョンを確認できます。
  * (サーバー v1.2+ が必要)
* **🗑️ ごみ箱機能**：
  * ノート削除後、自動的にごみ箱に移動させることができます。
  * ごみ箱からのノート復元をサポートします。(添付ファイルの復元機能は順次追加予定)

* **🚫 オフライン同期戦略**：
  * オフラインでのノート編集の自動マージをサポート。(プラグイン側での設定が必要)
  * オフラインでの削除に対し、再接続時に自動的に補完、あるいは削除同期を実行。(プラグイン側での設定が必要)

* **🔗 共有機能**：
  * ノートの共有の作成/取り消しが可能。
  * 共有ノート内で参照されている画像、音声、動画などの添付ファイルを自動で解析します。
  * 共有アクセスの統計機能を提供。
  * 共有ノートにパスワードを設定可能。
  * 共有ノートへの短縮URL（ショートリンク）を生成可能。
* **📂 ディレクトリ同期**：
  * フォルダの作成/名前変更/移動/削除の同期をサポート。

* **🌳 Gitの自動化**：
  * 添付ファイルやノートに変更があった際、自動的にリモートGitリポジトリへ更新およびプッシュを実行。
  * タスク終了後に自動的にシステムのメモリを解放。

* **☁️ マルチストレージバックアップと一方向ミラー同期**：
  * S3/OSS/R2/WebDAV/ローカル など、複数のストレージプロトコルに対応。
  * 全体/差分ZIPスケジュールアーカイブバックアップをサポート。
  * Vaultリソースのリモートストレージへの一方向ミラー同期をサポート。
  * 有効期限切れのバックアップの自動クリーンアップに対応し、保存期間のカスタマイズが可能。

* **🗄️ マルチデータベース対応**：
  * SQLite、MySQL、PostgreSQL など主流データベースをネイティブでサポートし、個人からチームまでの多様なデプロイニーズに応えます。

## ☕ スポンサーとサポート

- このプラグインが非常に役立つと感じ、今後も開発を継続してほしい場合は、以下の方法でサポートをご検討ください：

  | Ko-fi *中国以外の地域*                                                                               |    | WeChat QRコード *中国地域*                        |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | または | <img src="/docs/images/wxds.png" height="150"> |

  - サポートリスト：
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.ja.md">Support.ja.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.ja.md">Support.ja.md (cnb.cool ミラー)</a>

## ⏱️ 更新履歴 (Changelog)

- ♨️ [更新履歴を確認する](/docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ (Roadmap)

- [ ] 各階層を網羅する **Mock** テストの追加。
- [ ] WebSocketの `Protobuf` 転送フォーマットのサポート追加し、同期転送効率を強化。
- [ ] バックエンドに同期ログや操作ログなど、各種ログデータの参照機能を追加。
- [ ] 既存の認証メカニズムの分離および最適化を行い、全体的なセキュリティの向上を図る。
- [ ] WebGuiのノートのリアルタイム更新機能を追加。
- [ ] クライアント間のピアツーピアメッセージ転送機能の追加（ノートおよび添付ファイル以外。localsendのような機能。クライアント側での保存はサポートせず、サーバー側のみ保存可能）。
- [ ] 各種ヘルプドキュメントの充実。
- [ ] より多くの内向きネットワーク接続（中継ゲートウェイ）のサポート。
- [ ] 迅速なデプロイ計画：
  * サーバーアドレス（パブリックネットワーク）とアカウント、パスワードを提供するだけでFNSサーバーのデプロイが完了する仕組みの構築。
- [ ] 現在のオフラインでのノートの自動統合アルゴリズムを最適化し、競合の解決メカニズムを追加。

継続的に改善を行っており、以下の将来の開発計画があります：

> **改善の提案や新しいアイデアがある場合は、issueを送信して私たちと共有してください。内容を慎重に評価し、適切な提案を採用させていただきます。**

## 🚀 迅速なデプロイ(Quick Deployment)

複数のインストール方法が提供されています。**ワンクリックインストールスクリプト** または **Docker** の使用を推奨します。

### 方法1：ワンクリックインストールスクリプト（推奨）

システム環境を自動検出し、インストールとサービスの登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

中国地域の場合は、Tencentの `cnb.cool` ミラーリポジトリを使用できます。
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**スクリプトの主な動作：**

  * 現在のシステムに最適なReleaseバイナリファイルを自動的にダウンロードします。
  * デフォルトで `/opt/fast-note` にインストールされ、`/usr/local/bin/fns` にグローバルなショートカットコマンド `fns` を生成します。
  * Systemd（Linux）または Launchd（macOS）のサービスを設定・起動し、PC起動時の自動起動を実現します。
  * **管理用コマンド**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **インタラクティブメニュー**: `fns` を直接実行することで、インタラクティブなメニュー画面を呼び出し、インストール/アップグレード、コントロール、自動起動設定、GitHub / CNBミラー間の切り替えなどをサポートします。

-----

### 方法2：Dockerでのデプロイ

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
      - "9000:9000"  # RESTful API & WebSocketポート（ /api/user/sync がWebSocketインターフェースのアドレスになります）
    volumes:
      - ./storage:/fast-note-sync/storage  # データストレージ領域
      - ./config:/fast-note-sync/config    # 設定ファイル領域
```

サービスを起動します：

```bash
docker compose up -d
```

-----

### 方法3：手動でのバイナリインストール

ご使用のシステムに対応する最新バージョンを [Releases](https://github.com/haierkeys/fast-note-sync-service/releases) からダウンロードし、解凍して以下を実行します：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 ご利用ガイド

1.  **管理パネルへのアクセス**：
    ブラウザで `http://{サーバーIP}:9000` にアクセスします。
2.  **初期設定**：
    初回アクセス時はアカウントの登録が必要です。*(登録機能をオフにしたい場合は、設定ファイルで `user.register-is-enable: false` に設定してください)*
3.  **クライアントの設定**：
    管理パネルにログインし、**「API設定をコピー(Copy API Configuration)」** をクリックします。
4.  **Obsidianへの接続**：
    Obsidianのプラグイン設定画面を開き、先ほどコピーした設定情報を貼り付けて適用します。


## ⚙️ 設定に関する説明

デフォルトの設定ファイル「`config.yaml`」は、プログラムによって **ルートディレクトリ** または **config/** ディレクトリで自動的に検索されます。

完全な設定の例を確認する: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginxリバースプロキシ設定の例

完全な設定の例を確認する: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) サポート

FNSは **MCP (Model Context Protocol)** をネイティブサポートしています。

FNSをMCPサーバーとして、Cherry Studio、Cursorなどの互換性のあるAIクライアントに直接接続できます。接続後、AIはプライベートノートや添付ファイルの読み書き能力を備えます。さらに、MCPから発生したすべての変更は、WebSocketを通じてリアルタイムで各デバイス端末に同期されます。

### アクセス設定 (SSEモード)

FNSは **SSEプロトコル** を通じてMCPインターフェースを提供します。一般的なパラメータの要件は次のとおりです：
- **インターフェースアドレス**: `http://<あなたのサーバーIPまたはドメイン>:<ポート>/api/mcp/sse`
- **認証Header**: `Authorization: Bearer <あなたのAPIトークン>`（WebGUIの「API設定をコピー」機能から取得できます）
- **オプションのHeader**: `X-Default-Vault-Name: <ノート Vault の名前>`（MCP 操作でのデフォルトのノート Vault を指定するために使用されます。ツール呼び出し時に `vault` パラメータが指定されていない場合は、これが使用されます）
- **オプションのHeader**: `X-Client: <クライアントの種類>`（MCPへの接続に使用するクライアントの種類。例：Cherry Studio / OpenClaw）
- **オプションのHeader**: `X-Client-Version: <クライアントのバージョン>`（MCPへの接続に使用するクライアントのバージョン。例：1.1）
- **オプションのHeader**: `X-Client-Name: <クライアント名>`（MCPへの接続に指定されたクライアント名。例：Mac）



#### 例：Cherry Studio / Cursor / Cline など

ご自身のMCPクライアント設定にて、以下の記述をご参考ください：
*(注： `<ServerIP>`、`<Port>`、`<Token>`、および `<VaultName>` を各自の実際の情報に置き換えてください)*

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

## 🔗 クライアントとクライアントプラグイン

* Obsidian Fast Note Sync プラグイン
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool ミラー](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* サードパーティクライアント
  * [FastNodeSync-CLI ](https://github.com/Go1c/FastNodeSync-CLI) PythonおよびFNS WS APIを利用した双方向リアルタイム同期コマンドラインクライアント。GUIを持たないLinuxサーバー（OpenClawなど）向けに特化しており、Obsidianデスクトップやモバイル版に相当する完全な同期能力を提供します。
