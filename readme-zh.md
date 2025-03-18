[中文文档](readme-zh.md) / [English Document](README.md)
# Obsidian Better Sync Service

<p align="center">
    <img src="https://img.shields.io/github/release/haierkeys/obsidian-better-sync-service" alt="version">
    <img src="https://img.shields.io/github/license/haierkeys/obsidian-better-sync-service" alt="license">
</p>

该项目为 [Custom Image Auto Uploader](https://github.com/haierkeys/obsidian-custom-image-auto-uploader) Obsidian 插件提供图片上传、存储与云同步服务。

## 功能清单

- [x] 支持图片上传
- [x] 支持令牌授权，提升 API 安全性
- [x] 支持图片 HTTP 访问（基础功能，建议使用 Nginx 替代）
- [x] 存储支持：
  - [x] 同时保存至本地或云存储，方便后续迁移
  - [x] 本地存储支持（为 NAS 准备，功能已测试）
  - [x] 支持阿里云 OSS 云存储（功能已实现，尚未测试）
  - [x] 支持 Cloudflare R2 云存储（功能已实现，已测试）
  - [x] 支持 Amazon S3（功能已实现，已测试）
  - [x] 增加 MinIO 存储支持（v1.5+）
  - [ ] 支持 Google ECS（待开发）
- [x] 提供 Docker 安装支持，便于在家庭 NAS 或远程服务器上使用
- [x] 提供公共服务 API && Web界面，方便提供公共服务 <a href="#userapi">用户公共接口 & Web 界面</a>

## 更新日志

查看完整的更新内容，请访问 [Changelog](https://github.com/haierkeys/obsidian-better-sync-service/releases)。

## 价格

本软件是开源且免费的。如果您想表示感谢或帮助支持继续开发，可以通过以下方式为我提供支持：

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 快速开始
### 安装

- 目录设置

  ```bash
  # 创建项目所需的目录
  mkdir -p /data/image-api
  cd /data/image-api

  mkdir -p ./config && mkdir -p ./storage/logs && mkdir -p ./storage/uploads
  ```

  首次启动如果不下载配置文件,程序会自动生成一个默认配置到 **config/config.yaml**

  如果你想从网络下载一个默认配置 使用以下命令来下载

  ```bash
  # 从开源库下载默认配置文件到配置目录
  wget -P ./config/ https://raw.githubusercontent.com/haierkeys/obsidian-better-sync-service/main/config/config.yaml
  ```

- 二进制安装

  从 [Releases](https://github.com/haierkeys/obsidian-better-sync-service/releases) 下载最新版本，解压后执行：

  ```bash
  ./image-api run -c config/config.yaml
  ```


- 容器化安装（Docker 方式）

  Docker 命令:

  ```bash
  # 拉取最新的容器镜像
  docker pull haierkeys/obsidian-better-sync-service:latest

  # 创建并启动容器
  docker run -tid --name image-api \
          -p 9000:9000 -p 9001:9001 \
          -v /data/image-api/storage/:/api/storage/ \
          -v /data/image-api/config/:/api/config/ \
          haierkeys/obsidian-better-sync-service:latest
  ```

  Docker Compose
  使用 *containrrr/watchtower* 来监听镜像实现自动更新项目
  **docker-compose.yaml** 内容如下

  ```yaml
  # docker-compose.yaml
  services:
    image-api:
      image: haierkeys/obsidian-better-sync-service:latest  # 你的应用镜像
      container_name: image-api
      ports:
        - "9000:9000"  # 映射端口 9000
        - "9001:9001"  # 映射端口 9001
      volumes:
        - /data/image-api/storage/:/api/storage/  # 映射存储目录
        - /data/image-api/config/:/api/config/    # 映射配置目录

    watchtower:
      image: containrrr/watchtower
      container_name: watchtower
      volumes:
        - /var/run/docker.sock:/var/run/docker.sock  # 允许 Watchtower 访问 Docker Daemon
      environment:
        - WATCHTOWER_SCHEDULE=0 0,30 * * * *  # 每半小时检查更新
        - WATCHTOWER_CLEANUP=true            # 删除旧镜像，节省空间
      restart: unless-stopped
  ```

  执行 **docker compose**

  以服务方式注册 docker 容器

  ```bash
  docker compose up -d
  ```

  注销并销毁 docker 容器

  ```bash
  docker compose down
  ```



### 使用

- **使用单服务网关**

	支持 `本地存储`, `OSS` , `Cloudflare R2` , `Amazon S3` , `MinIO`

	需要修改 [config.yaml](config/config.yaml#http-port)

	修改`http-port` 和 `auth-token` 两个选项

	启动网关程序

	API 网关地址为 `http://{IP:PORT}/api/upload`

	API 访问令牌为  `auth-token` 内容


- **使用 多用户 开放网关**

	支持  `本地存储`, `OSS` , `Cloudflare R2` , `Amazon S3` , `MinIO`  ( v2.3+ )

	需要在 [config.yaml](config/config.yaml#user) 中修改

	`http-port` 和 `database`

	同时修改 `user.is-enable` 和 `user.register-is-enable` 为 `true`

	启动网关程序

	访问 `WebGUI` 地址 `http://{IP:PORT}` 进行用户注册配置

	![Image](https://github.com/user-attachments/assets/39c798de-b243-42c1-a75a-cd179913fc49)

	API 网关地址为 `http://{IP:PORT}/api/user/upload`

	点击在 `WebGUI` 复制 API 配置 获取配置信息



- **存储类型说明**


  | 存储类型       | 说明 |
  |----------------|-----------|
  | 服务器本地存储   | 默认的保存路径为: `/data/storage/uploads` 关联配置项`config.local-fs.save-path` 为 `storage/uploads`,  <br />如果使用网关图片资源访问服务, 需要 `config.local-fs.httpfs-is-enable` 设置为 `true` <br /> 对应的 `访问地址前缀` 为 `http://{IP:PORT}`, 使用单服务网关设置 `config.app.upload-url-pre` <br />推荐使用 Nginx 来实现资源访问 |



### 配置说明

默认的配置文件名为 **config.yaml**，请将其放置在 **根目录** 或 **config** 目录下。

更多配置详情请参考：

- [config/config.yaml](config/config.yaml)


## 其他资源

- [Obsidian Auto Image Remote Uploader](https://github.com/haierkeys/obsidian-auto-image-remote-uploader)