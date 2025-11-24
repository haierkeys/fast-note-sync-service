[中文文档](readme-zh.md) / [English Document](README.md)
# Fast Note Sync Service

<p align="center">
    <img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service" alt="version">
    <img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service" alt="license">
</p>



基于 Golang + Websocket + Sqlite + React 构建的高性能 低延迟 笔记同步服务 , 需要和 [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) 配合使用

----

##### WebGui:

笔记仓库：<br />

<img src="https://image.diybeta.com/blog/fast-note-sync-service-1.png" alt="fast-note-sync-service-2" width="800" />

笔记管理：<br />

<img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-2" width="800" />


## 功能清单

- [x] 多端笔记实时同步
- [ ] 笔记云存储同步备份 - s3
- [ ] 笔记云存储同步备份 - 阿里云
- [ ] 笔记云存储同步备份 - CF R2
- [ ] 笔记云存储同步备份 - minio
- [ ] 笔记云存储同步备份 - webdav
- [ ] 笔记云存储同步备份 - 增加备份策略
- [x] Web页面管理
- [x] 目前仅支持 Sqlite 存储
- [x] 增加仓库管理
- [x] 增加笔记管理 （ 支持 新建/删除/重命名/查看/编辑 笔记 ，所有更改会实时同步到所有设备 ）
- [ ] 增加git维护版本
- [ ] 基于 google-diff-match-patch 算法优化


## 更新日志

查看完整的更新内容，请访问 [Changelog](https://github.com/haierkeys/fast-note-sync-service/releases)。

## 价格

本软件是开源且免费的。如果您想表示感谢或帮助支持继续开发，可以通过以下方式为我提供支持：

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 私有部署

1) 推荐：一键安装脚本（自动检测平台并安装）

  ```bash
  bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/quest_install.sh)
  ```
脚本主要功能：
- 自动下载对应平台的 Release 资产并安装二进制
- 创建 /opt/fast-note 作为默认安装目录
- 在 /usr/local/bin/fast-note 创建快捷命令
- 为 systemd 环境创建 systemd 服务（fast-note.service）并启用
- 支持命令：install|uninstall|full-uninstall|start|stop|status|update|install-self|menu
- full-uninstall 会删除安装目录、日志、安装脚本与软链（执行前有交互确认）

2) 二进制手动安装

  从 [Releases](https://github.com/haierkeys/fast-note-sync-service/releases) 下载最新版本，解压后执行：

  ```bash
  ./fast-note-sync-service run -c config/config.yaml
  ```

3) 容器化安装（Docker 命令方式）

  Docker 命令:

  ```bash
  # 拉取最新的容器镜像
  docker pull haierkeys/fast-note-sync-service:latest

  # 创建并启动容器
  docker run -tid --name fast-note-sync-service \
          -p 9000:9000 -p 9001:9001 \
          -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
          -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
          haierkeys/fast-note-sync-service:latest
  ```

4) 容器化安装（Docker Compose）
  **docker-compose.yaml** 内容如下

  ```yaml
  # /data/docker-compose.yaml
  services:
    fast-note-sync-service:
      image: haierkeys/fast-note-sync-service:latest
      container_name: fast-note-sync-service
      ports:
        - "9000:9000"
        - "9001:9001"
      volumes:
        - /data/fast-note-sync/storage/:/fast-note-sync/storage/  # 映射存储目录
        - /data/fast-note-sync/config/:/fast-note-sync/config/    # 映射配置目录
      networks:
        - app-network  # 与 image-api 在同一网络1

  ```

  执行 **docker compose**

  以服务方式注册 docker 容器

  ```bash
  docker compose -f /data/docker-compose.yaml up -d
  ```

  注销并销毁 docker 容器

  ```bash
 docker compose -f /data/docker-compose.yaml down
  ```

### 使用

访问 `WebGUI` 地址 `http://{IP:PORT}`

点击在 复制 API 配置 获取配置信息, 到 `Fast Note Sync For Obsidian` 插件中粘贴即可

首次访问需要进行用户注册,如需关闭注册, 请修改 `user.register-is-enable` 为 `false`





### 配置说明

默认的配置文件名为 **config.yaml**，请将其放置在 **根目录** 或 **config** 目录下。

更多配置详情请参考：

- [config/config.yaml](config/config.yaml)


## 其他资源

- [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync)