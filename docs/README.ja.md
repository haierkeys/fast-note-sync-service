[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

ご不明な点がございましたら、新しい [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) を作成するか、Telegram 交流グループに参加して助けを求めてください: [https://t.me/obsidian_users](https://t.me/obsidian_users)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能、低遅延のノート同期、オンライン管理、リモート REST API などのサービスプラットフォーム</strong>
  <br>
  <em>Golang + Websocket + Sqlite + React ベースで構築</em>
</p>

<p align="center">
  データの提供にはクライアントプラグインが必要です：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
  <div align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/vault.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/vault.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/attach.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/attach.png" alt="fast-note-sync-service-preview" width="400" /></a>
    </div>
  <div align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/note.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/note.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/setting.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/setting.png" alt="fast-note-sync-service-preview" width="400" /></a>
  </div>
</div>

---

## ✨ コア機能

* **🚀 REST API 対応**:
    * 標準的な REST API インターフェースを提供し、プログラミング方式（自動化スクリプト、AI アシスタント統合など）による Obsidian ノートの CRUD 操作をサポートします。
    * 詳細は [REST API ドキュメント](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/REST_API.md) を参照してください。
* **💻 Web 管理パネル**:
  * モダナイズされた管理インターフェースを内蔵し、ユーザー作成、プラグイン設定の生成、リポジトリおよびノート内容の管理が簡単に行えます。
* **🔄 マルチデバイス・ノート同期**:
    * **Vault (リポジトリ)** の自動作成をサポート。
    * ノート管理（追加、削除、変更、検索）をサポートし、変更はミリ秒単位でリアルタイムにすべてのオンラインデバイスに配信されます。
* **🖼️ 添付ファイル同期対応**:
    * 画像などのノート以外のファイルの同期を完璧にサポート。
    * 大きな添付ファイルの分割アップロードとダウンロードをサポート。分割サイズは設定可能で、同期効率を向上させます。
* **⚙️ 設定の同期**:
    * `.obsidian` 設定ファイルの同期をサポート。
    * `PDF` の閲覧進捗状態の同期をサポート。
* **📝 ノート履歴**:
    * Web ページまたはプラグイン側で、各ノートの過去の修正バージョンを確認できます。
    * (サーバー v1.2+ が必要です)
* **🗑️ ゴミ箱**:
    * ノートの削除後、自動的にゴミ箱に移動するようにサポート。
    * ゴミ箱からのノート復元をサポート。（今後、順次添付ファイルの復旧機能も追加予定）

* **🚫 オフライン同期戦略**:
    * ノートのオフライン編集時の自動マージをサポート。（プラグイン側の設定が必要）
    * オフラインでの削除、再接続後の自動補完または削除同期をサポート。（プラグイン側の設定が必要）

## ☕ 寄付とサポート

- このプラグインが有用だと感じ、開発を継続してほしい場合は、以下の方法で私をサポートしてください：

  | Ko-fi *中国国外*                                                                                                     |     | WeChat 投げ銭 *中国国内*                                            |
  |----------------------------------------------------------------------------------------------------------------------|-----|--------------------------------------------------------------------|
  | [<img src="https://ik.imagekit.io/haierkeys/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | または | <img src="https://ik.imagekit.io/haierkeys/wxds.png" height="150"> |

  - 寄付者リスト：
    - https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md

## ⏱️ 更新履歴

- ♨️ [更新履歴を表示](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ (Roadmap)

継続的に改善を行っています。以下は今後の開発計画です：

- [ ] **共有機能**: ノートの共有をサポート。
- [ ] **MCP 対応**: AI MCP 関連機能のサポートを追加。
- [ ] **ディレクトリ同期**: ディレクトリの CRUD 操作をサポート。
- [ ] **Git バージョン管理の統合**: ノートのより安全なバージョン履歴遡及を提供。
- [ ] **クラウドストレージとバックアップ戦略**:
    - [ ] カスタムバックアップ戦略の設定。
    - [ ] マルチプロトコル対応：S3 / Minio / Cloudflare R2 / Alibaba Cloud OSS / WebDAV。

> **改善の提案や新しいアイデアがありましたら、issue を通じて共有してください。適切な提案は慎重に評価し、採用させていただきます。**

## 🚀 クイックデプロイ

複数のインストール方法を提供しています。**ワンクリック・スクリプト** または **Docker** の使用を推奨します。

### 方法1：ワンクリック・スクリプト（推奨）

システム環境を自動的に検出し、インストールとサービスの登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**スクリプトの主な動作：**

  * 現在のシステムに適した Release バイナリファイルを自動的にダウンロードします。
  * デフォルトで `/opt/fast-note` にインストールされ、`/usr/local/bin/fast-note` にショートカットコマンドが作成されます。
  * Systemd サービス (`fast-note.service`) を設定して起動し、OS 起動時の自動開始を実現します。
  * **管理コマンド**: `fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### 方法2：Docker デプロイ

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

サービスを開始します：

```bash
docker compose up -d
```

-----

### 方法3：手動バイナリインストール

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) から対応するシステムの最新バージョンをダウンロードし、解凍して実行します：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 ユーザーガイド

1.  **管理パネルへのアクセス**:
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**:
    初回アクセス時にはアカウントの登録が必要です。*(登録機能を無効にするには、設定ファイルで `user.register-is-enable: false` と設定してください)*
3.  **クライアントの設定**:
    管理パネルにログインし、**「API 設定をコピー」**をクリックします。
4.  **Obsidian との接続**:
    Obsidian のプラグイン設定ページを開き、コピーした設定情報を貼り付ければ完了です。


## ⚙️ 設定の説明

デフォルトの設定ファイルは `config.yaml` です。プログラムは**ルートディレクトリ**または **config/** ディレクトリ内を自動的に検索します。

完全な設定例を表示：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx リバースプロキシ設定例

完全な設定例を表示：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 関連リソース

  * [Obsidian Fast Note Sync Plugin (クライアントプラグイン)](https://github.com/haierkeys/obsidian-fast-note-sync)
