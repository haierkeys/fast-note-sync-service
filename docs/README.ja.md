[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>高性能・低遅延のノート同期サービスソリューション</strong>
  <br>
  <em>Golang + Websocket + Sqlite + React で構築</em>
</p>

<p align="center">
  クライアントプラグインと併用する必要があります：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ 主な機能

* **💻 Web管理パネル**：モダンな管理インターフェースを内蔵し、ユーザー作成、プラグイン設定の生成、リポジトリおよびノートの管理を簡単に行えます。
* **🔄 マルチデバイス同期**：
    * **Vault (リポジトリ)** の自動作成をサポート。
    * ノート管理（追加、削除、変更、検索）をサポートし、変更はミリ秒単位でリアルタイムにすべてのオンラインデバイスに配信されます。
* **🖼️ 添付ファイルの同期サポート**：
    * 画像などの非ノートファイルの同期を完全にサポート。
    * *(注：サーバー v0.9+ および [Obsidian プラグイン v1.0+](https://github.com/haierkeys/obsidian-fast-note-sync/releases) が必要です。Obsidian の設定ファイルはサポートしていません)*
* **📝 ノート履歴**：
    * Webページまたはプラグイン側で、各ノートの過去の修正バージョンを確認できます。
    * (サーバー v1.2+ が必要)
* **⚙️ 設定の同期**：
    * `.obsidian` 設定ファイルの同期をサポート。

## ⏱️ 更新履歴

- ♨️ [更新履歴を確認する](docs/CHANGELOG.ja.md)

## 🗺️ ロードマップ (Roadmap)

継続的に改善を行っており、以下の開発計画を予定しています：

- [ ] **Git バージョン管理の統合**：ノートのより安全なバージョン履歴を提供。
- [ ] **同期アルゴリズムの最適化**：`google-diff-match-patch` を統合し、より効率的な増分同期を実現。
- [ ] **クラウドストレージとバックアップ戦略**：
    - [ ] カスタムバックアップ戦略の設定。
    - [ ] マルチプロトコル対応：S3 / Minio / Cloudflare R2 / Alibaba Cloud OSS / WebDAV。

> **改善の提案や新しいアイデアがございましたら、issue を通じてお気軽にお知らせください。適切な提案は慎重に検討し、採用させていただきます。**

## 🚀 クイックデプロイ

複数のインストール方法を提供していますが、**一クリックスクリプト** または **Docker** の使用を推奨します。

### 方法1：一クリックスクリプト（推奨）

システム環境を自動的に検出し、インストールとサービス登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**スクリプトの主な動作：**

  * 現在のシステムに適した Release バイナリファイルを自動的にダウンロード。
  * デフォルトで `/opt/fast-note` にインストールし、`/usr/local/bin/fast-note` にショートカットコマンドを作成。
  * Systemd サービス (`fast-note.service`) を設定・起動し、OS 起動時の自動実行を実現。
  * **管理コマンド**：`fast-note [install|uninstall|start|stop|status|update|menu]`

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

`docker-compose.yaml` ファイルを作成：

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

### 方法3：手動バイナリインストール

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) から対応するシステムの最新バージョンをダウンロードし、解凍して実行してください：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用ガイド

1.  **管理パネルへのアクセス**：
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**：
    初回アクセス時にアカウント登録が必要です。*(登録機能をオフにする場合は、設定ファイルで `user.register-is-enable: false` を設定してください)*
3.  **クライアントの設定**：
    管理パネルにログインし、**「API 設定をコピー」**をクリックします。
4.  **Obsidian との接続**：
    Obsidian のプラグイン設定画面を開き、コピーした設定情報を貼り付けてください。

## ⚙️ 設定について

デフォルトの設定ファイルは `config.yaml` です。プログラムは自動的に **ルートディレクトリ** または **config/** ディレクトリ内を検索します。

完全な設定例を確認する：[config/config.yaml](https://www.google.com/search?q=config/config.yaml)

## 📅 更新履歴

完全なバージョン更新記録については、[Releases ページ](https://github.com/haierkeys/fast-note-sync-service/releases) をご覧ください。

## ☕ 支援とスポンサー

このプロジェクトは完全にオープンソースで無料です。もしお役に立てましたら、プロジェクトへの **Star** や著者へのコーヒー一杯の支援をお願いいたします。継続的なメンテナンスの励みになります。ありがとうございます！

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 🔗 関連リソース

  * [Obsidian Fast Note Sync Plugin (クライアントプラグイン)](https://github.com/haierkeys/obsidian-fast-note-sync)
