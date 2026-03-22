[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

何か問題があれば [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) を作成するか、Telegram グループに参加して助けを求めてください: [https://t.me/obsidian_users](https://t.me/obsidian_users)

中国本土では、Tencent の `cnb.cool` ミラーリポジトリの使用を推奨します: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能、低遅延のノート同期、オンライン管理、リモート REST API サービスプラットフォーム</strong>
  <br>
  <em>Golang + Websocket + Sqlite + React ベース</em>
</p>

<p align="center">
  データ提供はクライアントプラグインと併用する必要があります：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## ✨ コア機能

* **🚀 REST API 対応**:
    * 標準的な REST API インターフェースを提供し、プログラム（自動化スクリプト、AI アシスタントの統合など）を介した Obsidian ノートの作成、読み取り、更新、削除をサポートします。
    * 詳細は [RESTful API ドキュメント](/docs/REST_API.md) または [OpenAPI ドキュメント](/docs/swagger.yaml) を参照してください。
* **💻 Web 管理パネル**:
  * モダンな管理インターフェースを内蔵し、ユーザー作成、プラグイン設定の生成、リポジトリおよびノート内容の管理を簡単に行えます。
* **🔄 マルチデバイス同期**:
    * **Vault (倉庫)** の自動作成をサポート。
    * ノート管理（追加、削除、変更、検索）をサポート。変更はミリ秒単位でリアルタイムにすべてのオンラインデバイスに配信されます。
* **🖼️ 添付ファイル同期対応**:
    * 画像などの非ノートファイルの同期を完璧にサポート。
    * 大容量添付ファイルの分割アップロードとダウンロードをサポート（分割サイズ設定可能）し、同期効率を高めます。
* **⚙️ 設定同期**:
    * `.obsidian` 設定ファイルの同期をサポート。
    * `PDF` 進捗ステータスの同期をサポート。
* **📝 ノート履歴**:
    * Web ページまたはプラグイン側から各ノートの修正履歴を確認できます。
    * (サーバー v1.2+ が必要)
* **🗑️ ゴミ箱**:
    * ノート削除後、自動的にゴミ箱に移動します。
    * ゴミ箱からのノート復元をサポート。（添付ファイルの復元機能は順次追加予定）

* **🚫 オフライン同期戦略**:
    * オフライン編集の自動マージをサポート（プラグイン側の設定が必要）。
    * オフライン削除、再接続後の自動補完または削除同期（プラグイン側の設定が必要）。

* **🔗 共有機能**:
    * ノート共有の作成/解除が可能です。
    * 共有ノート内で参照されている画像、音声、動画などの添付ファイルを自動的に解析します。
    * 共有アクセスの統計機能を提供します。
* **📂 ディレクトリ同期**:
    * フォルダの作成/名前変更/移動/削除の同期をサポート。

* **🌳 Git 自動化**:
    * 添付ファイルやノートが変更された際、リモート Git リポジトリへ自動的に更新・プッシュします。
    * タスク終了後にシステムメモリを自動的に解放します。

* **☁️ マルチストレージバックアップと一方向ミラー同期**:
    * S3/OSS/R2/WebDAV/ローカルなど、多様なストレージプロトコルに対応。
    * 全量/増量 ZIP 定期アーカイブバックアップをサポート。
    * Vault リソースのリモートストレージへの一方向ミラー同期をサポート。
    * 期限切れバックアップの自動クリーンアップ（保存日数のカスタマイズ可能）をサポート。

## ☕ スポンサーとサポート

- このプラグインが便利だと感じ、開発を継続してほしいと思われる方は、以下の方法でサポートをお願いします：

  | Ko-fi *中国以外*                                                                                 |    | WeChat で寄付 *中国*                           |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - 支援者リスト：
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.ja.md">Support.ja.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.ja.md">Support.ja.md (cnb.cool ミラー)</a>

## ⏱️ 更新履歴

- ♨️ [更新履歴を確認する](/docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ (Roadmap)

継続的に改善を行っています。今後の開発計画は以下の通りです：


- [ ] **🤖 MCP 対応**: AI MCP 関連機能のサポート追加。
- [ ] **より多くのデータベースタイプのサポート**

> **改善の提案や新しいアイデアがあれば、issue を通じて共有してください。適切な提案は慎重に評価し、採用させていただきます。**

## 🚀 クイックデプロイ

多様なインストール方法を提供していますが、**ワンクリックスクリプト** または **Docker** を推奨します。

### 方法 1：ワンクリックスクリプト（推奨）

システム環境を自動的に検出し、インストールとサービス登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

中国では Tencent の `cnb.cool` ミラー源を使用できます：
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**スクリプトの主な動作：**

  * 現在のシステムに適した Release バイナリファイルを自動的にダウンロードします。
  * デフォルトで `/opt/fast-note` にインストールし、`/usr/local/bin/fns` にグローバルショートカットコマンド `fns` を作成します。
  * Systemd (Linux) または Launchd (macOS) サービスを設定して起動し、自動起動を実現します。
  * **管理コマンド**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **対話型メニュー**: `fns` を直接実行すると、インストール/アップグレード、サービス制御、自動起動設定、GitHub / CNB ミラーの切り替えが可能な対話型メニューに入ります。

-----

### 方法 2：Docker デプロイ

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
      - "9000:9000"  # RESTful API & WebSocket ポート。 /api/user/sync が WebSocket インターフェースアドレスです。
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

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) から対応するシステムの最新バージョンをダウンロードし、解凍後に実行してください：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 ユーザーガイド

1.  **管理パネルへのアクセス**:
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**:
    初回アクセス時にアカウント登録が必要です。*(登録機能をオフにするには、設定ファイルで `user.register-is-enable: false` を設定してください)*
3.  **クライアントの設定**:
    管理パネルにログインし、「**API 設定をコピー**」をクリックします。
4.  **Obsidian との接続**:
    Obsidian のプラグイン設定ページを開き、コピーした設定情報を貼り付けます。


## ⚙️ 設定説明

デフォルトの設定ファイルは `config.yaml` です。プログラムは**ルートディレクトリ**または **config/** ディレクトリ内を自動的に検索します。

完全な設定例を表示：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx リバースプロキシ設定例

完全な設定例を表示：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 クライアント & プラグイン

* Obsidian Fast Note Sync プラグイン
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool ミラー](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* サードパーティクライアント
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) Python と FNS WS インターフェースに基づき実装された、高性能な双方向リアルタイム同期コマンドラインクライアント。GUI のない Linux サーバー環境（OpenClaw など）に適しており、Obsidian デスクトップ/モバイルと同等の同期能力を実現します。
