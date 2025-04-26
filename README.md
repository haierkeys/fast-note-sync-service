[中文文档](readme-zh.md) / [English Document](README.md)
# Better Sync Service

<p align="center">
    <img src="https://img.shields.io/github/release/haierkeys/obsidian-better-sync-service" alt="version">
    <img src="https://img.shields.io/github/license/haierkeys/obsidian-better-sync-service" alt="license">
</p>

[BetterSync For Obsidian](https://github.com/haierkeys/obsidian-better-sync) server, a high-performance note real-time synchronization service built on Golang + Websocket.


## Feature List

- [x] Real-time synchronization of notes across multiple devices
- [ ] Note cloud storage synchronization
- [x] Web page management
- [x] Currently only supports Sqlite storage


## Changelog

For a complete list of updates, please visit [Changelog](https://github.com/haierkeys/obsidian-better-sync-service/releases).

## Pricing

This software is open-source and free. If you wish to express your gratitude or help support continued development, you can support me in the following ways:

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## Private Deployment

- Directory Setup

  ```bash
  # Create the directories required for the project
  mkdir -p /data/better-sync
  cd /data/better-sync

  mkdir -p ./config && mkdir -p ./storage/logs && mkdir -p ./storage/uploads
  ```

  If the configuration file is not downloaded at the first startup, the program will automatically generate a default configuration to **config/config.yaml**.

  If you want to download a default configuration from the network, use the following command to download it.

  ```bash
  # Download the default configuration file from the open-source repository to the configuration directory
  wget -P ./config/ https://raw.githubusercontent.com/haierkeys/obsidian-better-sync-service/main/config/config.yaml
  ```

- Binary Installation

  Download the latest version from [Releases](https://github.com/haierkeys/obsidian-better-sync-service/releases), extract it, and execute:

  ```bash
  ./better-sync-service run -c config/config.yaml
  ```


- Containerized Installation (Docker method)

  Docker command:

  ```bash
  # Pull the latest container image
  docker pull haierkeys/obsidian-better-sync-service:latest

  # Create and start the container
  docker run -tid --name better-sync-service \
          -p 9000:9000 -p 9001:9001 \
          -v /data/better-sync/storage/:/better-sync/storage/ \
          -v /data/better-sync/config/:/better-sync/config/ \
          haierkeys/obsidian-better-sync-service:latest
  ```

  Docker Compose
  Use *containrrr/watchtower* to monitor the image for automatic project updates
  The **docker-compose.yaml** content is as follows

  ```yaml
  # docker-compose.yaml
  services:
    better-sync:
      image: haierkeys/obsidian-better-sync-service:latest  # Your application image
      container_name: better-sync
      ports:
        - "9000:9000"  # Map port 9000
        - "9001:9001"  # Map port 9001
      volumes:
        - /data/better-sync/storage/:/better-sync/storage/  # Map storage directory
        - /data/better-sync/config/:/better-sync/config/    # Map configuration directory

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

Click to copy API configuration to get the configuration information, then paste it into the `BetterSync For Obsidian` plugin.

The first visit requires user registration. To disable registration, please change `user.register-is-enable` to `false`.


### Configuration Instructions

The default configuration file is named **config.yaml**, please place it in the **root directory** or **config** directory.

For more configuration details, please refer to:

- [config/config.yaml](config/config.yaml)


## Other Resources

- [Better Sync For Obsidian](https://github.com/haierkeys/obsidian-better-sync)