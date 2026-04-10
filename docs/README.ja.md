[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

問題がある場合は [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) を作成するか、Telegramグループに参加してください: [https://t.me/obsidian_users](https://t.me/obsidian_users)

中国本土のユーザーには、Tencent `cnb.cool` ミラーの利用を推奨します: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)



<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能・低レイテンシのノート同期、オンライン管理、リモート REST API サービスプラットフォーム</strong>
  <br>
  <em>Golang + WebSocket + SQLite + React で構築</em>
</p>

<p align="center">
  データ提供にはクライアントプラグインが必要です：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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
  * `FNS` を MCP サーバーとして `Cherry Studio`、`Cursor` などの互換 AI クライアントに接続することで、AI がプライベートノートや添付ファイルを読み書きできるようになり、すべての変更が WebSocket を通じてリアルタイムで各デバイス端末に同期されます。
* **🚀 REST API サポート**：
  * 標準的な REST API インターフェースを提供し、プログラムによるアクセス（自動化スクリプト、AI アシスタント連携など）で Obsidian ノートの CRUD 操作が可能です。
  * 詳細は [RESTful API ドキュメント](/docs/REST_API.md) または [OpenAPI ドキュメント](/docs/swagger.yaml) を参照してください。
* **💻 Web 管理パネル**：
  * モダンな管理インターフェースを内蔵。ユーザーの作成、プラグイン設定の生成、Vault・ノートの管理が簡単に行えます。
* **🔄 マルチデバイス ノート同期**：
  * **Vault（保管庫）** の自動作成をサポート。
  * ノート管理（追加・削除・更新・検索）をサポートし、変更はミリ秒単位で全オンラインデバイスにリアルタイム配信されます。
* **🖼️ 添付ファイル同期サポート**：
  * 画像などのノート以外のファイルの同期を完全サポート。
  * 大容量添付ファイルのチャンク アップロード/ダウンロードをサポート。チャンクサイズは設定可能で、同期効率を向上させます。
* **⚙️ 設定同期**：
  * `.obsidian` 設定ファイルの同期をサポート。
  * `PDF` の閲覧進捗状態の同期をサポート。
* **📝 ノート履歴**：
  * Web ページおよびプラグイン側で、各ノートの過去の変更バージョンを確認できます。
  * （サーバー v1.1+ が必要）
* **🗑️ ごみ箱**：
  * 削除されたノートは自動的にごみ箱へ移動します。
  * ごみ箱からノートを復元できます。（添付ファイルの復元機能は今後追加予定）

* **🚫 オフライン同期戦略**：
  * オフライン編集されたノートの自動マージをサポート。（プラグイン側の設定が必要）
  * オフライン削除は、再接続後に自動的に補完または同期されます。（プラグイン側の設定が必要）

* **🔗 共有機能**：
  * ノートの共有を作成/取り消しできます。
  * 共有ノート内で参照されている画像・音声・動画などの添付ファイルを自動解析します。
  * 共有アクセス統計機能を提供します。
  * 共有ノートにアクセスパスワードを設定できます。
  * 共有ノートの短縮リンクを生成できます。
* **📂 ディレクトリ同期**：
  * フォルダの作成/名前変更/移動/削除の同期をサポート。

* **🌳 Git 自動化**：
  * 添付ファイルやノートに変更が生じると、自動的にリモート Git リポジトリへ更新・プッシュします。
  * タスク完了後、システムメモリを自動解放します。

* **☁️ マルチストレージ バックアップ & 一方向ミラー同期**：
  * S3/OSS/R2/WebDAV/ローカルなど複数のストレージプロトコルに対応。
  * フル/増分 ZIP 定期アーカイブバックアップをサポート。
  * Vault リソースのリモートストレージへの一方向ミラー同期をサポート。
  * 期限切れバックアップの自動クリーンアップ。保持日数はカスタマイズ可能。

* **🗄️ マルチデータベース対応**:
  * SQLite、MySQL、PostgreSQL など、主要なデータベースをネイティブにサポート。個人からチームまで、さまざまなデプロイメントのニーズに対応します。

## ☕ スポンサー & サポート

- このプロジェクトが役に立ち、継続的な開発をサポートしたい場合は、以下の方法で支援をお願いします：

  | Ko-fi *（中国以外）*                                                                              |    | WeChat Pay *（中国）*                          |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | または | <img src="/docs/images/wxds.png" height="150"> |

  - サポーター一覧：
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md (cnb.cool ミラー)</a>

## ⏱️ 更新履歴

- ♨️ [更新履歴を見る](/docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ

- [ ] 各レイヤーをカバーする **Mock** テストの追加。
- [ ] WebSocket の **Protobuf** 転送フォーマットをサポートし、同期効率を向上。
- [ ] 同期ログや操作ログなど、各種ログのクエリ機能をバックエンドに追加。
- [ ] 既存の認可メカニズムの分離・最適化による全体的なセキュリティの向上。
- [ ] WebGUI でのノートのリアルタイム更新機能の追加。
- [ ] クライアント間の P2P メッセージング機能の追加（ノートや添付ファイル以外、LocalSend のような機能。クライアント側での保存は不可、サーバー側での保存は可能）。
- [ ] 各種ヘルプドキュメントの充実。
- [ ] イントラネット貫通（リレーゲートウェイ）のさらなるサポート。
- [ ] クイックデプロイメント計画
  * サーバーアドレス（公衆）、アカウント、パスワードのみで FNS サーバーのデプロイが完了するように。
- [ ] 既存のオフラインノート統合案の最適化と競合解決メカニズムの追加。

継続的に改善中です。以下が今後の開発計画です：

> **改善のご提案や新しいアイデアがあれば、issue を通じてぜひ共有してください。適切な提案は真剣に評価・採用いたします。**

## 🚀 クイックデプロイ

複数のインストール方法を提供しています。**ワンクリックスクリプト** または **Docker** を推奨します。

### 方法 1：ワンクリックスクリプト（推奨）

システム環境を自動検出し、インストールとサービス登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

中国のユーザーは Tencent `cnb.cool` ミラーを使用できます：
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**スクリプトの主な動作：**

  * 現在のシステムに対応する Release バイナリを自動ダウンロード。
  * デフォルトで `/opt/fast-note` にインストールし、`/usr/local/bin/fns` にグローバルショートカットコマンド `fns` を作成。
  * Systemd（Linux）または Launchd（macOS）サービスを設定・起動し、起動時に自動起動します。
  * **管理コマンド**：`fns [install|uninstall|start|stop|status|update|menu]`
  * **インタラクティブメニュー**：`fns` を直接実行するとインタラクティブメニューが表示され、インストール/アップグレード、サービス制御、起動設定、GitHub/CNB ミラーの切り替えが可能です。

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

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) からお使いのシステムの最新バージョンをダウンロードし、解凍して実行：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用ガイド

1.  **管理パネルへアクセス**：
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**：
    初回アクセス時にアカウントを登録します。*（登録を無効にするには、設定ファイルで `user.register-is-enable: false` を設定してください）*
3.  **クライアントの設定**：
    管理パネルにログインし、**「API 設定をコピー」** をクリックします。
4.  **Obsidian に接続**：
    Obsidian プラグイン設定ページを開き、コピーした設定情報を貼り付けます。


## ⚙️ 設定説明

デフォルトの設定ファイルは `config.yaml` です。プログラムは **ルートディレクトリ** または **config/** ディレクトリを自動的に検索します。

完全な設定例を確認：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx リバースプロキシ設定例

完全な設定例を確認：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP（モデルコンテキストプロトコル）サポート

FNS は **MCP (Model Context Protocol)** をネイティブサポートします。

`FNS` を MCP サーバーとして `Cherry Studio`、`Cursor` などの互換 AI クライアントに接続することで、AI がプライベートノートや添付ファイルを読み書きできるようになり、すべての変更が WebSocket を通じてリアルタイムで各デバイス端末に同期されます。

### 接続設定（SSE モード）

FNS は **SSE プロトコル** を通じて MCP インターフェースを提供します。一般的なパラメータは以下の通りです：
- **エンドポイント URL**：`http://<サーバーIPまたはドメイン>:<ポート>/api/mcp/sse`
- **認証ヘッダー**：`Authorization: Bearer <API トークン>`（WebGUI の「API 設定をコピー」から取得）
- **オプションのヘッダー**：`X-Default-Vault-Name: <VaultName>`（ツール呼び出しで `vault` パラメータが指定されていない場合に、MCP 操作のデフォルト Vault を指定するために使用）


#### 例：Cherry Studio / Cursor / Cline など

お使いの MCP クライアントで以下の設定を参考にしてください：
*（注：`<ServerIP>`、`<Port>`、`<Token>`、`<VaultName>` を実際の情報に置き換えてください）*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp/sse",
      "type": "sse",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer <Token>",
        "X-Default-Vault-Name": "<VaultName>"
      }
    }
  }
}
```

## 🔗 クライアント & クライアントプラグイン

* Obsidian Fast Note Sync プラグイン
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool ミラー](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* サードパーティクライアント
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) — Python と FNS WS インターフェースに基づく双方向リアルタイム同期コマンドラインクライアント。GUI のない Linux サーバー環境（OpenClaw など）に適しており、Obsidian デスクトップ/モバイルと同等の同期機能を実現します。
