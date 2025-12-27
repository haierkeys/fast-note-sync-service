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
  クライアント側のプラグインが必要です：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ コア機能

* **💻 Web管理パネル**：ユーザーの作成、プラグイン設定の生成、リポジトリやノートコンテンツの管理を簡単に行えるモダンな管理インターフェースを内蔵。
* **🔄 マルチデバイスノート同期**：
    * **Vault（倉庫）**の自動作成をサポート。
    * ノート管理（追加、削除、変更、検索）をサポートし、変更をミリ秒単位ですべてのオンラインデバイスにリアルタイム配信。
* **🖼️ 添付ファイル同期サポート**：
    * 画像などの非ノートファイルの同期を完全にサポート。
    * *(注：サーバー v0.9+ および [Obsidian プラグイン v1.0以上](https://github.com/haierkeys/obsidian-fast-note-sync/releases) が必要です。Obsidianの設定ファイルには対応していません)*
* **📝 ノート履歴**：
    * Webページやプラグイン側で、各ノートの変更履歴バージョンを表示できます。
    * (サーバー v1.2+ が必要)
* **⚙️ 設定同期**：
    * `.obsidian` 設定ファイルの同期をサポート。


## 🗺️ ロードマップ (Roadmap)

私たちは継続的に改善を行っています。今後の開発計画は以下の通りです：

- [ ] **Gitバージョン管理の統合**：ノートのより安全なバージョンバックトラックを提供。
- [ ] **同期アルゴリズムの最適化**：`google-diff-match-patch` を統合し、より効率的な増分同期を実現。
- [ ] **クラウドストレージとバックアップ戦略**：
    * [ ] カスタムバックアップ戦略設定。
    * [ ] マルチプロトコル対応：S3 / Minio / Cloudflare R2 / Aliyun OSS / WebDAV。

> **改善の提案や新しいアイデアがある場合は、Issueを提出して共有してください。私たちは慎重に評価し、適切な提案を採用します。**

## 🚀 迅速なデプロイ

複数のインストール方法を提供していますが、**ワンクリックスクリプト** または **Docker** の使用を推奨します。

### 方法1：ワンクリックスクリプト（推奨）

システム環境を自動的に検出し、インストールとサービス登録を完了します。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**スクリプトの主な動作：**

  * 現在のシステムに適したReleaseバイナリフォルダを自動的にダウンロード。
  * デフォルトで `/opt/fast-note` にインストールし、`/usr/local/bin/fast-note` にショートカットコマンドを作成。
  * 起動時に自動開始するように Systemd サービス (`fast-note.service`) を設定して開始。
  * **管理コマンド**：`fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### 方法2：Docker デプロイ

#### Docker Run

```bash
# 1. イメージをプル
docker pull haierkeys/fast-note-sync-service:latest

# 2. コンテナを起動
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

サービスを起動：

```bash
docker compose up -d
```

-----

### 方法3：手動バイナリインストール

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases) からお使いのシステムに対応した最新バージョンをダウンロードし、解凍して実行してください：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 使用ガイド

1.  **管理パネルへのアクセス**：
    ブラウザで `http://{サーバーIP}:9000` を開きます。
2.  **初期設定**：
    初回アクセス時にはアカウント登録が必要です。*(登録機能を無効にする場合は、設定ファイルで `user.register-is-enable: false` に設定してください)*
3.  **クライアントの設定**：
    管理パネルにログインし、**「API設定をコピー」** をクリックします。
4.  **Obsidian の接続**：
    Obsidian プラグインの設定ページを開き、コピーした設定情報を貼り付けます。

## ⚙️ 設定説明

デフォルトの設定ファイルは `config.yaml` です。プログラムは **ルートディレクトリ** または **config/** ディレクトリ内を自動的に検索します。

完全な設定例を表示：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 📅 更新履歴

完全なバージョン履歴を確認するには、[Releases ページ](https://github.com/haierkeys/fast-note-sync-service/releases) にアクセスしてください。

## ☕ 支援とサポート

このプロジェクトは完全にオープンソースで無料です。もしお役に立った場合は、このプロジェクトに **Star** を付けるか、作者にコーヒーを一杯ご馳走していただけると、継続的なメンテナンスの励みになります。ありがとうございます！

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 🔗 関連リソース

  * [Obsidian Fast Note Sync Plugin (クライアントプラグイン)](https://github.com/haierkeys/obsidian-fast-note-sync)
