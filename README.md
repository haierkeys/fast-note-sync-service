[中文文档](readme-zh.md) / [English Document](README.md)
# Obsidian Better Sync Service

<p align="center">
    <img src="https://img.shields.io/github/release/haierkeys/obsidian-better-sync-service" alt="version">
    <img src="https://img.shields.io/github/license/haierkeys/obsidian-better-sync-service" alt="license">
</p>

This project provides image upload, storage, and cloud synchronization services for the [Custom Image Auto Uploader](https://github.com/haierkeys/obsidian-custom-image-auto-uploader) Obsidian plugin.

## Feature List

- [x] Support for image uploads
- [x] Support for token authorization to enhance API security
- [x] Support for HTTP image access (basic functionality, it is recommended to use Nginx instead)
- [x] Storage options:
  - [x] Save to both local storage or cloud storage for easy migration in the future
  - [x] Local storage support (tested for NAS use)
  - [x] Support for Aliyun OSS cloud storage (implemented, yet to be tested)
  - [x] Support for Cloudflare R2 cloud storage (implemented and tested)
  - [x] Support for Amazon S3 (implemented and tested)
  - [x] Added support for MinIO storage (v1.5+)
  - [ ] Support for Google ECS (under development)
- [x] Provides Docker installation support to facilitate use on home NAS or remote servers
- [x] Provides public service API and Web interface, convenient for providing public services <a href="#userapi">User Public Interfaces & Web Interface</a>

## Update Log

For details on updates, please visit the [Changelog](https://github.com/haierkeys/obsidian-better-sync-service/releases).

## Pricing

This software is open source and free. If you want to express your gratitude or support further development, you can provide support through the following methods:

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## Quick Start

### Installation

- Directory Setup

  ```bash
  # Create directories needed for the project
  mkdir -p /data/image-api
  cd /data/image-api

  mkdir -p ./config && mkdir -p ./storage/logs && mkdir -p ./storage/uploads
  ```

  If you do not download the configuration file on the first startup, the program will automatically generate a default configuration at **config/config.yaml**.

  If you want to download a default configuration from the web, use the following command:

  ```bash
  # Download the default configuration file from the open-source library to the configuration directory
  wget -P ./config/ https://raw.githubusercontent.com/haierkeys/obsidian-better-sync-service/main/config/config.yaml
  ```

- Binary Installation

  Download the latest version from [Releases](https://github.com/haierkeys/obsidian-better-sync-service/releases), extract it, and run:

  ```bash
  ./image-api run -c config/config.yaml
  ```


- Containerized Installation (Using Docker)

  Docker Command:

  ```bash
  # Pull the latest container image
  docker pull haierkeys/obsidian-better-sync-service:latest

  # Create and start the container
  docker run -tid --name image-api \
          -p 9000:9000 -p 9001:9001 \
          -v /data/image-api/storage/:/api/storage/ \
          -v /data/image-api/config/:/api/config/ \
          haierkeys/obsidian-better-sync-service:latest
  ```

  Docker Compose:
  Use *containrrr/watchtower* to monitor the image and automatically update the project.
  Content of the **docker-compose.yaml**:

  ```yaml
  # docker-compose.yaml
  services:
    image-api:
      image: haierkeys/obsidian-better-sync-service:latest  # Your application image
      container_name: image-api
      ports:
        - "9000:9000"  # Map port 9000
        - "9001:9001"  # Map port 9001
      volumes:
        - /data/image-api/storage/:/api/storage/  # Map storage directory
        - /data/image-api/config/:/api/config/    # Map configuration directory

    watchtower:
      image: containrrr/watchtower
      container_name: watchtower
      volumes:
        - /var/run/docker.sock:/var/run/docker.sock  # Allow Watchtower to access Docker Daemon
      environment:
        - WATCHTOWER_SCHEDULE=0 0,30 * * * *  # Check for updates every 30 minutes
        - WATCHTOWER_CLEANUP=true            # Remove old images to save space
      restart: unless-stopped
  ```

  Execute **docker compose**

  To register the docker container as a service:

  ```bash
  docker compose up -d
  ```

  To unregister and destroy the docker container:

  ```bash
  docker compose down
  ```


### Usage

- **Using Single Service Gateway**

  Supports `Local Storage`, `OSS`, `Cloudflare R2`, `Amazon S3`, `MinIO`

  Modify [config.yaml](config/config.yaml#http-port)

  Modify `http-port` and `auth-token` options

  Start the gateway program

  The API gateway address is `http://{IP:PORT}/api/upload`

  The API access token is the content of `auth-token`


- **Using Multi-user Public Gateway**

  Supports `Local Storage`(v2.3+), `OSS`, `Cloudflare R2`, `Amazon S3`, `MinIO` (v2.3+)

  Modify in [config.yaml](config/config.yaml#user)

  `http-port` and `database`

  Also modify `user.is-enable` and `user.register-is-enable` to `true`

  Start the gateway program

  Access the `WebGUI` address `http://{IP:PORT}` to register and configure users

  ![Image](https://github.com/user-attachments/assets/39c798de-b243-42c1-a75a-cd179913fc49)

  The API gateway address is `http://{IP:PORT}/api/user/upload`

  Click to copy API configuration in the `WebGUI` to obtain configuration information


- **Storage Type Description**


| Storage Type         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
|----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Server Local Storage | The default save path is: `/data/storage/uploads` related to the configuration item `config.local-fs.save-path` is `storage/uploads`, <br /> If using the gateway image resource access service, it is necessary to set `config.local-fs.httpfs-is-enable` to `true` <br /> The corresponding `Access Address Prefix` is `http://{IP:PORT}`, using the single service gateway setting `config.app.upload-url-pre` <br /> Recommended to use Nginx to achieve resource access |

### Configuration Description

The default configuration file name is **config.yaml**, please place it in the **root directory** or **config** directory.

For more details, please refer to:

- [config/config.yaml](config/config.yaml)


## Additional Resources

- [Obsidian Auto Image Remote Uploader](https://github.com/haierkeys/obsidian-auto-image-remote-uploader)

