[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能・低遅延なノート同期サービスソリューション</strong>
  <br>
  <em>Golang + Websocket + Sqlite + React で構築</em>
</p>

<p align="center">
  クライアントプラグインと一緒に使用する必要があります：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ 核心機能

* **💻 Web 管理パネル**: モダンな管理インターフェースを内蔵し、ユーザー作成、プラグイン設定の生成、リポジトリおよびノート内容の管理を簡単に行えます。
* **🔄 マルチデバイス同期**:
    * **Vault (リポジトリ)** の自動作成をサポート。
    * ノート管理（追加、削除、編集、検索）をサポート。変更はミリ秒単位でオンラインのすべてのデバイスにリアルタイムに配信されます。
* **🖼️ 添付ファイルの同期**:
    * 画像などの非ノートファイルの同期を完全にサポート。
    * *(注：サーバー v0.9+ および [Obsidian プラグイン v1.0+ ](https://github.com/haierkeys/obsidian-fast-note-sync/releases) が必要です。Obsidian の設定ファイルはサポートされていません)*
* **📝 ノート履歴**:
    * Web ページやプラグイン端から、各ノートの過去の変更履歴を確認できます。
    * (サーバー v1.2+ が必要)
* **⚙️ 設定同期**:
    * `.obsidian` 設定ファイルの同期をサポートしています。

## ☕ 支援とスポンサー

- このプラグインが有用だと感じ、開発を継続してほしい場合は、以下の方法で私を支援してくださると幸いです。

  | Ko-fi *中国以外*  |  | 微信 (WeChat) スキャンで寄付 *中国* |
  | --- | ---| --- |
  | [<img src="https://ik.imagekit.io/haierkeys/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | または | <img src="https://ik.imagekit.io/haierkeys/wxds.png" height="150"> |

## ⏱️ 更新履歴

- ♨️ [更新履歴を表示](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ (Roadmap)

継続的に改善を行っています。今後の開発計画は以下の通りです。

- [ ] **Git バージョン管理の統合**: ノートのより安全なバージョン履歴の遡及を提供します。
- [ ] **同期アルゴリズムの最適化**: `google-diff-match-patch` を統合し、より効率的な増分同期を実現します。
- [ ] **クラウドストレージとバックアップ戦略**:
    - [ ] カスタムバックアップ戦略の設定。
    - [ ] マルチプロトコル対応：S3 / Minio / Cloudflare R2 / Aliyun OSS / WebDAV。

> **改善の提案や新しいアイデアがある場合は、issue を通じて共有してください。適切な提案を慎重に評価し、採用します。**

## 🚀 クイックデプロイ

様々なインストール方法を提供していますが、**ワンクリックスクリプト** または **Docker** を推奨します。

### 方法1：ワンクリックスクリプト（推奨）

システム環境を自動検出し、インストールとサービス登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**スクリプトの主な動作：**

  * 現在のシステムに適した Release バイナリファイルを自動的にダウンロードします。
  * デフォルトで `/opt/fast-note` にインストールし、`/usr/local/bin/fast-note` にショートカットを作成します。
  * Systemd サービス (`fast-note.service`) を設定して起動し、OS 起動時の自動実行を実現します。
  * **管理コマンド**: `fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### 方法2：Docker デプロイ

#### Docker Run

```bash
# 1. イメージをプルする
docker pull haierkeys/fast-note-sync-service:latest

# 2. コンテナを起動する
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

サービスを起動します：

```bash
docker compose up -d
```

-----

### 方法3：手動バイナリインストール

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) から対応するシステムの最新バージョンをダウンロードし、解凍して実行します：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用ガイド

1.  **管理パネルへのアクセス**:
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**:
    初回アクセス時にアカウント登録が必要です。*(登録機能をオフにする場合は、設定ファイルで `user.register-is-enable: false` を設定してください)*
3.  **クライアントの設定**:
    管理パネルにログインし、「**API 設定をコピー**」をクリックします。
4.  **Obsidian への接続**:
    Obsidian プラグインの設定ページを開き、今コピーした設定情報を貼り付けます。

## ⚙️ 設定の説明

デフォルトの設定ファイルは `config.yaml` で、プログラムは **ルートディレクトリ** または **config/** ディレクトリから自動的に検索します。

完全な設定例を表示：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)


## 🔗 関連リソース

  * [Obsidian Fast Note Sync Plugin (クライアントプラグイン)](https://github.com/haierkeys/obsidian-fast-note-sync)
