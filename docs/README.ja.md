[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

ご質問がある場合は、新しい [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) を作成するか、Telegram グループに参加して助けを求めてください: [https://t.me/obsidian_users](https://t.me/obsidian_users)

中国大陸地区では、Tencent `cnb.cool` ミラーリポジトリの使用をお勧めします: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


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
  <em>Golang + Websocket + Sqlite + React で構築</em>
</p>

<p align="center">
  データ提供にはクライアントプラグインとの併用が必要です：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## ✨ 主な機能

* **🚀 REST API 対応**:
    * 標準的な REST API インターフェースを提供し、プログラム（自動化スクリプト、AI アシスタント統合など）を介した Obsidian ノートの CRUD 操作をサポートします。
    * 詳細は [REST API ドキュメント](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/REST_API.md) を参照してください。
* **💻 Web 管理パネル**:
  * モダンな管理インターフェースを内蔵し、ユーザー作成、プラグイン設定の生成、リポジトリおよびノート内容の管理が簡単に行えます。
* **🔄 マルチデバイスノート同期**:
    * **Vault (リポジトリ)** の自動作成をサポート。
    * ノート管理（追加、削除、修正、検索）をサポートし、変更はミリ秒単位でリアルタイムにすべてのオンラインデバイスに配布されます。
* **🖼️ 添付ファイル同期対応**:
    * 画像などのノート以外のファイルの同期を完全にサポート。
    * 大きな添付ファイルのチャンクアップロード・ダウンロードをサポートし、チャンクサイズを設定可能で同期効率を向上させます。
* **⚙️ 設定同期**:
    * `.obsidian` 設定ファイルの同期をサポート。
    * `PDF` 進捗ステータスの同期をサポート。
* **📝 ノート履歴**:
    * Web ページやプラグイン側で、各ノートの過去の修正バージョンを確認できます。
    * (サーバー v1.2+ が必要)
* **🗑️ ゴミ箱**:
    * ノート削除後、自動的にゴミ箱に移動します。
    * ゴミ箱からのノート復元をサポート。（今後、添付ファイルの復元機能も順次追加予定）

* **🚫 オフライン同期戦略**:
    * オフライン編集時の自動マージをサポート（プラグイン側の設定が必要）。
    * オフライン削除、再接続後の自動同期または削除をサポート（プラグイン側の設定が必要）。

## ☕ スポンサーとサポート

- このプラグインが便利だと感じ、開発の継続をサポートしたい場合は、以下の方法でご支援をお願いします：

  | Ko-fi *中国国外*                                                                                 |    | WeChat Pay *中国国内*                          |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - 支援者リスト：
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.ja.md">Support.ja.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.ja.md">Support.ja.md (cnb.cool ミラー)</a>

## ⏱️ 更新履歴

- ♨️ [更新履歴を表示](/docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ (Roadmap)

継続的に改善を行っています。今後の開発計画は以下の通りです：

- [ ] **共有機能**: ノートの共有をサポート。
- [ ] **MCP 対応**: AI MCP 関連機能のサポートを追加。
- [ ] **ディレクトリ同期**: ディレクトリの CRUD 操作をサポート。
- [ ] **Git バージョン管理統合**: ノートのより安全なバージョン追跡を提供。
- [ ] **クラウドストレージとバックアップ戦略**:
    - [ ] カスタムバックアップ戦略設定。
    - [ ] マルチプロトコル対応：S3 / Minio / Cloudflare R2 / Aliyun OSS / WebDAV。

> **改善の提案や新しいアイデアがある場合は、issue を送信して共有してください。適切な提案を真剣に検討し、採用します。**

## 🚀 クイックデプロイ

複数のインストール方法を提供していますが、**ワンクリックスクリプト** または **Docker** の使用をお勧めします。

### 方法 1：ワンクリックスクリプト（推奨）

システム環境を自動的に検出し、インストールとサービス登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

中国地区では Tencent `cnb.cool` ミラーソースを使用できます：
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh?cnb)
```


**スクリプトの主な動作：**

  * 現在のシステムに適した Release バイナリファイルを自動的にダウンロードします。
  * デフォルトで `/opt/fast-note` にインストールされ、`/usr/local/bin/fast-note` にショートカットコマンドが作成されます。
  * Systemd サービス (`fast-note.service`) を設定・起動し、OS 起動時の自動実行を実現します。
  * **管理コマンド**: `fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### 方法 2：Docker デプロイ

#### Docker Run

```bash
# 1. イメージのプル
docker pull haierkeys/fast-note-sync-service:latest

# 2. コンテナの起動
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 -p 9001:9001 \
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
      - "9000:9000"  # API ポート
      - "9001:9001"  # WebSocket ポート
    volumes:
      - ./storage:/fast-note-sync/storage  # データストレージ
      - ./config:/fast-note-sync/config    # 設定ファイル
```

サービスの起動：

```bash
docker compose up -d
```

-----

### 方法 3：手動バイナリインストール

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) からシステムに対応した最新バージョンをダウンロードし、解凍して実行します：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用ガイド

1.  **管理パネルへのアクセス**:
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**:
    初回アクセス時にアカウント登録が必要です。*(登録機能を無効にするには、設定ファイルで `user.register-is-enable: false` を設定してください)*
3.  **クライアントの設定**:
    管理パネルにログインし、**「API 設定をコピー」** をクリックします。
4.  **Obsidian との接続**:
    Obsidian のプラグイン設定ページを開き、コピーした設定情報を貼り付けます。


## ⚙️ 設定の説明

デフォルトの設定ファイルは `config.yaml` です。プログラムは自動的に **ルートディレクトリ** または **config/** ディレクトリ内を検索します。

完全な設定例を表示：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx リバースプロキシ設定例

完全な設定例を表示：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 関連リソース

  * [Obsidian Fast Note Sync Plugin (クライアントプラグイン)](https://github.com/haierkeys/obsidian-fast-note-sync)
  * [Obsidian Fast Note Sync Plugin (cnb.cool ミラー)](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
