[中文文档](readme-zh.md) / [English Document](README.md)
# Fast Note Sync Service

<p align="center">
    <img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service" alt="version">
    <img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service" alt="license">
</p>

[Fast Note Sync Service for Obsidian](https://github.com/haierkeys/fast-note-sync-service) server, a high-performance note real-time synchronization service built on Golang + Websocket.


## Feature List

- [x] Real-time synchronization of notes across multiple devices
- [ ] Note cloud storage synchronization
- [x] Web page management
- [x] Currently only supports Sqlite storage


## Changelog

For a complete list of updates, please visit [Changelog](https://github.com/haierkeys/fast-note-sync-service/releases).

## Pricing

This software is open-source and free. If you wish to express your gratitude or help support continued development, you can support me in the following ways:

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## Private Deployment

- Directory Setup

  ```bash
  # Create the directories required for the project
  mkdir -p /data/fast-note-sync
  cd /data/fast-note-sync

  mkdir -p ./config && mkdir -p ./storage/logs && mkdir -p ./storage/uploads
  ```

  If the configuration file is not downloaded at the first startup, the program will automatically generate a default configuration to **config/config.yaml**.

  If you want to download a default configuration from the network, use the following command to download it.

  ```bash
  # Download the default configuration file from the open-source repository to the configuration directory
  wget -P ./config/ https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/main/config/config.yaml
  ```

- Binary Installation

  Download the latest version from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), extract it, and execute:

  ```bash
  ./fast-note-sync-service run -c config/config.yaml
  ```


- Containerized Installation (Docker method)

  Docker command:

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

  Docker Compose
  Use *containrrr/watchtower* to monitor the image for automatic project updates
  The **docker-compose.yaml** content is as follows

  ```yaml
  # docker-compose.yaml
  services:
    fast-note-sync-service:
      image: haierkeys/fast-note-sync-service:latest
      container_name: fast-note-sync-service
      ports:
        - "9000:9000"
        - "9001:9001"
      volumes:
        - /data/fast-note-sync/storage/:/fast-note-sync/storage/  # Map storage directory
        - /data/fast-note-sync/config/:/fast-note-sync/config/    # Map configuration directory
      networks:
        - app-network
  ```

  Execute **docker compose**

  Register the docker container as a service

  ```bash
  docker compose up -d
  ```

  Log out and destroy the docker container

  ```bash
  docker compose down
  ```

### Usage

Access the `WebGUI` address `http://{IP:PORT}`

Click to copy API configuration to get the configuration information, then paste it into the `Fast Note Sync For Obsidian` plugin.

The first visit requires user registration. To disable registration, please change `user.register-is-enable` to `false`.


### Configuration Instructions

The default configuration file is named **config.yaml**, please place it in the **root directory** or **config** directory.

For more configuration details, please refer to:

- [config/config.yaml](config/config.yaml)


## Other Resources

- [Fast Note Sync For Obsidian](https://github.com/haierkeys/obsidian-fast-note-sync)