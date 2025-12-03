[中文文档](readme-zh.md) / [English Document](README.md)
# Fast Note Sync Service

<p align="center">
    <img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service" alt="version">
    <img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service" alt="license">
</p>



A high-performance, low-latency note synchronization service built with Golang + WebSocket + SQLite + React. It needs to be used together with the [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync).


## Web GUI Screenshot:


Note repository: <br />

<img src="https://image.diybeta.com/blog/fast-note-sync-service-1.png" alt="fast-note-sync-service-2" width="800" />

Note management: <br />

<img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-2" width="800" />


## Features

- [x] Web page management
- [x] Currently only supports SQLite storage
- [x] Add repository management
- [x] Add note management ( Supports creating/deleting/renaming/viewing/editing notes, all changes will be synced to all devices in real time )
- [x] Real-time note sync across multiple devices
- [ ] Add support for images and other types of non-note attachments
- [ ] Add Git support
- [ ] Note cloud storage sync backup - S3
- [ ] Note cloud storage sync backup - Alibaba Cloud
- [ ] Note cloud storage sync backup - Cloudflare R2
- [ ] Note cloud storage sync backup - MinIO
- [ ] Note cloud storage sync backup - WebDAV
- [ ] Note cloud storage sync backup - Add backup strategies
- [ ] Optimize using the Google diff-match-patch algorithm1

## Changelog

To view the full changelog, please visit [Changelog](https://github.com/haierkeys/fast-note-sync-service/releases).

## Pricing

This software is open source and free. If you'd like to show appreciation or help support continued development, you can support me via:

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## Self-hosting

1) Recommended: one-click install script (automatically detects platform and installs)

  ```bash
  bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/quest_install.sh)
  ```
Main features of the script:
- Automatically download the Release asset for the detected platform and install the binary
- Create /opt/fast-note as the default installation directory
- Create a shortcut command at /usr/local/bin/fast-note
- Create and enable a systemd service (fast-note.service) for systemd environments
- Support commands: install|uninstall|full-uninstall|start|stop|status|update|install-self|menu
- full-uninstall will remove the installation directory, logs, install script, and symlink (interactive confirmation before execution)

2) Manual binary installation

  Download the latest release from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), extract and run:

  ```bash
  ./fast-note-sync-service run -c config/config.yaml
  ```

3) Containerized installation (Docker CLI)

  Docker commands:

  ```bash
  # Pull the latest container image
  docker pull haierkeys/fast-note-sync-service:latest

  # Create and start the container
  docker run -tid --name fast-note-sync-service \
          -p 9000:9000 -p 9001:9001 \
          -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
          -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
          haierkeys/fast-note-sync-service:latest
  ```

4) Containerized installation (Docker Compose)
  The **docker-compose.yaml** content is as follows

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
        - /data/fast-note-sync/storage/:/fast-note-sync/storage/  # map storage directory
        - /data/fast-note-sync/config/:/fast-note-sync/config/    # map config directory
      networks:
        - app-network  # on the same network as image-api

  ```

  Run **docker compose**

  Register the docker container as a service

  ```bash
  docker compose -f /data/docker-compose.yaml up -d
  ```

  Unregister and destroy the docker container

  ```bash
 docker compose -f /data/docker-compose.yaml down
  ```

### Usage

Visit the Web GUI at `http://{IP:PORT}`

Click "Copy API Configuration" to obtain the configuration info, then paste it into the `Fast Note Sync For Obsidian` plugin.

The first visit requires user registration. To disable registration, set `user.register-is-enable` to `false`.





### Configuration

The default configuration file name is **config.yaml**. Please place it in the **root directory** or the **config** directory.

For more configuration details, see:

- [config/config.yaml](config/config.yaml)


## Other resources

- [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync)