[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>고성능, 저지연 노트 동기화 서비스 솔루션</strong>
  <br>
  <em>Golang + Websocket + Sqlite + React 기반 구축</em>
</p>

<p align="center">
  클라이언트 플러그인과 함께 사용해야 합니다: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ 핵심 기능

* **💻 Web 관리 패널**: 현대적인 관리 인터페이스 내장, 사용자 생성, 플러그인 설정 생성, 저장소 및 노트 관리를 쉽게 할 수 있습니다.
* **🔄 멀티 디바이스 노트 동기화**:
    * **Vault (저장소)** 자동 생성 지원.
    * 노트 관리(추가, 삭제, 수정, 조회) 지원, 변경 사항은 밀리초 단위로 모든 온라인 디바이스에 실시간 배포됩니다.
* **🖼️ 첨부 파일 동기화 지원**:
    * 이미지 등 비노트 파일 동기화 완벽 지원.
    * *(참고: 서버 v0.9+ 및 [Obsidian 플러그인 v1.0+](https://github.com/haierkeys/obsidian-fast-note-sync/releases) 필요, Obsidian 설정 파일은 지원하지 않음)*
* **📝 노트 히스토리**:
    * 웹 페이지 및 플러그인 단에서 각 노트의 과거 수정 버전을 확인할 수 있습니다.
    * (서버 v1.2+ 필요)
* **⚙️ 설정 동기화**:
    * `.obsidian` 설정 파일의 동기화를 지원합니다.

## ⏱️ 업데이트 로그

- ♨️ [업데이트 로그 확인하기](docs/CHANGELOG.ko.md)

## 🗺️ 로드맵 (Roadmap)

지속적으로 개선 중이며, 향후 개발 계획은 다음과 같습니다:

- [ ] **Git 버전 관리 통합**: 노트에 대한 더 안전한 버전 회복 기능 제공.
- [ ] **동기화 알고리즘 최적화**: `google-diff-match-patch` 통합으로 더 효율적인 증분 동기화 구현.
- [ ] **클라우드 스토리지 및 백업 전략**:
    - [ ] 사용자 정의 백업 전략 설정.
    - [ ] 멀티 프로토콜 적응: S3 / Minio / Cloudflare R2 / Alibaba Cloud OSS / WebDAV.

> **개선 제안이나 새로운 아이디어가 있다면 issue를 통해 공유해 주세요. 적절한 제안은 신중히 검토하여 반영하겠습니다.**

## 🚀 빠른 배포

다양한 설치 방식을 제공하며, **원클릭 스크립트** 또는 **Docker** 사용을 권장합니다.

### 방식 1: 원클릭 스크립트 (권장)

시스템 환경을 자동으로 감지하고 설치 및 서비스 등록을 완료합니다.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**스크립트의 주요 동작:**

  * 현재 시스템에 적합한 Release 바이너리 파일을 자동으로 다운로드.
  * 기본적으로 `/opt/fast-note`에 설치되며, `/usr/local/bin/fast-note`에 바로가기 명령을 생성.
  * Systemd 서비스 (`fast-note.service`)를 구성 및 실행하여 부팅 시 자동 시작 구현.
  * **관리 명령**: `fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### 방식 2: Docker 배포

#### Docker Run

```bash
# 1. 이미지 풀
docker pull haierkeys/fast-note-sync-service:latest

# 2. 컨테이너 시작
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 -p 9001:9001 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

`docker-compose.yaml` 파일 생성:

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # API 포트
      - "9001:9001"  # WebSocket 포트
    volumes:
      - ./storage:/fast-note-sync/storage  # 데이터 저장소
      - ./config:/fast-note-sync/config    # 설정 파일
```

서비스 시작:

```bash
docker compose up -d
```

-----

### 방식 3: 수동 바이너리 설치

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases)에서 해당 시스템의 최신 버전을 다운로드하고 압축을 푼 후 실행하세요:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 사용 가이드

1.  **관리 패널 접속**:
    브라우저에서 `http://{서버 IP}:9000`을 엽니다.
2.  **초기 설정**:
    첫 방문 시 계정 등록이 필요합니다. *(등록 기능을 끄려면 설정 파일에서 `user.register-is-enable: false`로 설정하세요)*
3.  **클라이언트 설정**:
    관리 패널에 로그인한 후 **"API 설정 복사"**를 클릭합니다.
4.  **Obsidian 연결**:
    Obsidian 플러그인 설정 페이지를 열고 방금 복사한 설정 정보를 붙여넣으세요.

## ⚙️ 설정 설명

기본 설정 파일은 `config.yaml`이며, 프로그램은 자동으로 **루트 디렉토리** 또는 **config/** 디렉토리에서 검색합니다.

전체 설정 예시 확인: [config/config.yaml](https://www.google.com/search?q=config/config.yaml)

## 📅 업데이트 로그

전체 버전 업데이트 기록은 [Releases 페이지](https://github.com/haierkeys/fast-note-sync-service/releases)에서 확인하세요.

## ☕ 후원 및 지원

이 프로젝트는 완전한 오픈 소스이며 무료입니다. 도움이 되셨다면 프로젝트에 **Star**를 눌러주시거나 작성자에게 커피 한 잔을 후원해 주세요. 지속적인 유지 관리에 큰 힘이 됩니다. 감사합니다!

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 🔗 관련 리소스

  * [Obsidian Fast Note Sync Plugin (클라이언트 플러그인)](https://github.com/haierkeys/obsidian-fast-note-sync)
