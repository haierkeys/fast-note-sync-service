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
  클라이언트 플러그인이 필요합니다: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ 핵심 기능

* **💻 Web 관리 패널**: 내장된 모던 관리 인터페이스로 사용자 생성, 플러그인 설정 생성, 저장소 및 노트 콘텐츠 관리를 쉽게 할 수 있습니다.
* **🔄 멀티 디바이스 노트 동기화**:
    * **Vault (저장소)** 자동 생성을 지원합니다.
    * 노트 관리(추가, 삭제, 수정, 검색)를 지원하며, 변경 사항을 밀리초 단위로 모든 온라인 장치에 실시간 배포합니다.
* **🖼️ 첨부 파일 동기화 지원**:
    * 이미지 등 노트 외 파일 동기화를 완벽하게 지원합니다.
    * *(참고: 서버 v0.9+ 및 [Obsidian 플러그인 v1.0 이상](https://github.com/haierkeys/obsidian-fast-note-sync/releases) 필요; Obsidian 설정 파일 미지원)*
* **📝 노트 기록**:
    * 웹 페이지 및 플러그인 측에서 각 노트의 수정 기록 버전을 확인할 수 있습니다.
    * (서버 v1.2+ 필요)
* **⚙️ 설정 동기화**:
    * `.obsidian` 설정 파일 동기화를 지원합니다.


## 🗺️ 로드맵 (Roadmap)

지속적으로 개선 중입니다. 향후 개발 계획은 다음과 같습니다:

- [ ] **Git 버전 관리 통합**: 노트에 대해 더 안전한 버전 추적을 제공합니다.
- [ ] **동기화 알고리즘 최적화**: `google-diff-match-patch`를 통합하여 더 효율적인 증분 동기화를 구현합니다.
- [ ] **클라우드 스토리지 및 백업 전략**:
    * [ ] 사용자 정의 백업 전략 구성.
    * [ ] 멀티 프로토콜 지원: S3 / Minio / Cloudflare R2 / Aliyun OSS / WebDAV.

> **개선 제안이나 새로운 아이디어가 있다면 Issue를 제출하여 공유해 주세요. 신중하게 검토하고 적절한 제안을 채택하겠습니다.**

## 🚀 빠른 배포

다양한 설치 방법을 제공하며, **원클릭 스크립트** 또는 **Docker** 사용을 권장합니다.

### 방법 1: 원클릭 스크립트 (권장)

시스템 환경을 자동으로 감지하고 설치 및 서비스 등록을 완료합니다.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**스크립트 주요 동작:**

  * 현재 시스템에 맞는 Release 바이너리 파일을 자동으로 다운로드합니다.
  * 기본적으로 `/opt/fast-note`에 설치되며, `/usr/local/bin/fast-note`에 단축 명령이 생성됩니다.
  * 부팅 시 자동 시작되도록 Systemd 서비스(`fast-note.service`)를 구성하고 시작합니다.
  * **관리 명령**: `fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### 방법 2: Docker 배포

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
      - ./storage:/fast-note-sync/storage  # 데이터 스토리지
      - ./config:/fast-note-sync/config    # 구성 파일
```

서비스 시작:

```bash
docker compose up -d
```

-----

### 방법 3: 수동 바이너리 설치

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases)에서 시스템에 맞는 최신 버전을 다운로드하고 압축을 푼 후 실행합니다:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 사용 가이드

1.  **관리 패널 접속**:
    브라우저에서 `http://{서버IP}:9000`을 엽니다.
2.  **초기 설정**:
    첫 방문 시 계정 등록이 필요합니다. *(등록 기능을 비활성화하려면 설정 파일에서 `user.register-is-enable: false`로 설정하십시오)*
3.  **클라이언트 구성**:
    관리 패널에 로그인하고 **"API 구성 복사(Copy API Config)"**를 클릭합니다.
4.  **Obsidian 연결**:
    Obsidian 플러그인 설정 페이지를 열고 복사한 구성 정보를 붙여넣습니다.

## ⚙️ 구성 설명

기본 구성 파일은 `config.yaml`입니다. 프로그램은 자동으로 **루트 디렉터리** 또는 **config/** 디렉터리를 검색합니다.

전체 구성 예시 보기: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 📅 변경 로그

전체 버전 상세 기록을 보려면 [Releases 페이지](https://github.com/haierkeys/fast-note-sync-service/releases)를 방문하십시오.

## ☕ 후원 및 지원

이 프로젝트는 완전히 오픈 소스이며 무료입니다. 도움이 되었다면 프로젝트에 **Star**를 주시거나 저자에게 커피 한 잔을 사주세요. 지속적인 유지 관리에 큰 동기 부여가 됩니다. 감사합니다!

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 🔗 관련 리소스

  * [Obsidian Fast Note Sync Plugin (클라이언트 플러그인)](https://github.com/haierkeys/obsidian-fast-note-sync)
