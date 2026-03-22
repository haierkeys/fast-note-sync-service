[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

문제가 발생하면 [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new)를 생성하거나 Telegram 커뮤니티 그룹에 가입하여 도움을 요청하세요: [https://t.me/obsidian_users](https://t.me/obsidian_users)

중국 본토에서는 Tencent `cnb.cool` 미러 저장소를 사용하는 것을 권장합니다: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>고성능, 저지연 노트 동기화, 온라인 관리, 원격 REST API 서비스 플랫폼</strong>
  <br>
  <em>Golang + Websocket + Sqlite + React 기반 구축</em>
</p>

<p align="center">
  데이터 제공을 위해서는 클라이언트 플러그인이 필요합니다: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
  <div align="center">
    <a href="/docs/images/vault.png"><img src="/docs/images/vault.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="/docs/images/attach.png"><img src="/docs/images/attach.png" alt="fast-note-sync-service-preview" width="400" /></a>
    </div>
  <div align="center">
    <a href="/docs/images/note.png"><img src="/docs/images/note.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="/docs/images/setting.png"><img src="/docs/images/setting.png" alt="fast-note-sync-service-preview" width="400" /></a>
  </div>
</div>

---

## ✨ 핵심 기능

* **🚀 REST API 지원**:
    * 표준 REST API 인터페이스를 제공하여 프로그래밍 방식(자동화 스크립트, AI 비서 통합 등)으로 Obsidian 노트를 CRUD할 수 있도록 지원합니다.
    * 자세한 내용은 [RESTful API 문서](/docs/REST_API.md) 또는 [OpenAPI 문서](/docs/swagger.yaml)를 참조하세요.
* **💻 Web 관리 패널**:
  * 현대적인 관리 인터페이스를 내장하여 사용자 생성, 플러그인 설정 생성, 저장소 및 노트 콘텐츠 관리를 쉽게 할 수 있습니다.
* **🔄 멀티 디바이스 노트 동기화**:
    * **Vault (저장소)** 자동 생성 지원.
    * 노트 관리(추가, 삭제, 수정, 검색) 지원. 변경 사항은 모든 온라인 디바이스에 밀리초 단위로 실시간 배포됩니다.
* **🖼️ 첨부 파일 동기화 지원**:
    * 이미지 등 비노트 파일 동기화를 완벽하게 지원합니다.
    * 대용량 첨부 파일의 분할 업로드 및 다운로드를 지원(분할 크기 설정 가능)하여 동기화 효율을 높입니다.
* **⚙️ 설정 동기화**:
    * `.obsidian` 설정 파일 동기화를 지원합니다.
    * `PDF` 진행 상태 동기화를 지원합니다.
* **📝 노트 히스토리**:
    * 웹 페이지 또는 플러그인에서 각 노트의 수정 이력을 확인할 수 있습니다.
    * (서버 v1.2+ 필요)
* **🗑️ 휴지통**:
    * 노트 삭제 시 자동으로 휴지통으로 이동합니다.
    * 휴지통에서 노트를 복구할 수 있습니다. (첨부 파일 복구 기능은 추후 추가 예정)

* **🚫 오프라인 동기화 전략**:
    * 오프라인 편집 시 자동 병합을 지원합니다. (플러그인 설정 필요)
    * 오프라인 삭제 후 재연결 시 자동으로 동기화 보완 또는 삭제를 진행합니다. (플러그인 설정 필요)

* **🔗 공유 기능**:
    * 노트 공유를 생성하거나 취소할 수 있습니다.
    * 공유된 노트에 포함된 이미지, 오디오, 비디오 등의 첨부 파일을 자동으로 분석합니다.
    * 공유 접근 통계 기능을 제공합니다.
* **📂 디렉토리 동기화**:
    * 폴더 생성/이름 변경/이동/삭제 동기화를 지원합니다.

* **🌳 Git 자동화**:
    * 첨부 파일 및 노트 변경 시 원격 Git 저장소로 자동 업데이트 및 푸시를 진행합니다.
    * 작업 종료 후 시스템 메모리를 자동으로 해제합니다.

* **☁️ 멀티 스토리지 백업 및 단방향 미러 동기화**:
    * S3/OSS/R2/WebDAV/로컬 등 다양한 스토리지 프로토콜을 지원합니다.
    * 전체/증분 ZIP 정기 아카이브 백업을 지원합니다.
    * Vault 리소스를 원격 스토리지로 단방향 미러 동기화하는 기능을 지원합니다.
    * 만료된 백업 자동 정리 및 커스텀 보관 기간 설정을 지원합니다.

## ☕ 후원 및 지원

- 이 플러그인이 유용하며 지속적인 개발을 원하신다면 다음 방법을 통해 지원해 주세요:

  | Ko-fi *중국 외 지역*                                                                               |    | 위챗 페이 후원 *중국 지역*                          |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | 또는 | <img src="/docs/images/wxds.png" height="150"> |

  - 후원자 명단:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.ko.md">Support.ko.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.ko.md">Support.ko.md (cnb.cool 미러)</a>

## ⏱️ 업데이트 로그

- ♨️ [업데이트 로그 확인하기](/docs/CHANGELOG.ko.md)

## 🗺️ 로드맵 (Roadmap)

지속적으로 개선 중이며, 향후 개발 계획은 다음과 같습니다:


- [ ] **🤖 MCP 지원**: AI MCP 관련 기능 지원 추가.
- [ ] **더 많은 데이터베이스 유형 지원**

> **개선 제안이나 새로운 아이디어가 있다면 issue를 통해 공유해 주세요. 소중한 의견을 검토하여 반영하겠습니다.**

## 🚀 빠른 배포

다양한 설치 방법을 제공하며, **원클릭 스크립트** 또는 **Docker** 사용을 권장합니다.

### 방법 1: 원클릭 스크립트 (권장)

시스템 환경을 자동으로 감지하여 설치 및 서비스 등록을 완료합니다.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

중국 지역에서는 Tencent `cnb.cool` 미러를 사용할 수 있습니다:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**스크립트 주요 동작:**

  * 현재 시스템에 맞는 최신 Release 바이너리 파일을 자동으로 다운로드합니다.
  * 기본적으로 `/opt/fast-note`에 설치되며, `/usr/local/bin/fns`에 전역 명령 `fns`를 생성합니다.
  * Systemd(Linux) 또는 Launchd(macOS) 서비스를 구성하여 부팅 시 자동 실행을 설정합니다.
  * **관리 명령**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **대화형 메뉴**: `fns`를 실행하여 설치/업그레이드, 서비스 제어, 자동 실행 설정, GitHub / CNB 미러 전환 등을 수행할 수 있습니다.

-----

### 방법 2: Docker 배포

#### Docker Run

```bash
# 1. 이미지 풀
docker pull haierkeys/fast-note-sync-service:latest

# 2. 컨테이너 실행
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

`docker-compose.yaml` 파일을 생성합니다:

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # RESTful API & WebSocket 포트. /api/user/sync가 WebSocket 인터페이스 주소입니다.
    volumes:
      - ./storage:/fast-note-sync/storage  # 데이터 저장소
      - ./config:/fast-note-sync/config    # 설정 파일
```

서비스 시작:

```bash
docker compose up -d
```

-----

### 방법 3: 수동 바이너리 설치

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases)에서 해당 시스템의 최신 버전을 다운로드하고 압축을 푼 뒤 실행하세요:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 사용 가이드

1.  **관리 패널 접속**:
    브라우저에서 `http://{서버_IP}:9000`을 엽니다.
2.  **초기 설정**:
    첫 접속 시 계정 등록이 필요합니다. *(등록 기능을 끄려면 설정 파일에서 `user.register-is-enable: false`로 설정하세요)*
3.  **클라이언트 설정**:
    관리 패널에 로그인하여 **"API 설정 복사"**를 클릭합니다.
4.  **Obsidian 연결**:
    Obsidian 플러그인 설정 페이지에서 복사한 설정 정보를 붙여넣습니다.


## ⚙️ 설정 설명

기본 설정 파일은 `config.yaml`이며, 프로그램은 **루트 디렉토리** 또는 **config/** 디렉토리에서 자동으로 검색합니다.

전체 설정 예시 보기: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx 역방향 프록시 설정 예시

전체 설정 예시 보기: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 클라이언트 및 플러그인

* Obsidian Fast Note Sync 플러그인
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool 미러](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* 제3자 클라이언트
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) Python 및 FNS WS 인터페이스를 기반으로 구현된 고성능 양방향 실시간 동기화 CLI 클라이언트로, GUI가 없는 Linux 서버 환경(OpenClaw 등)에 적합하며 Obsidian 데스크톱/모바일과 동등한 동기화 성능을 제공합니다.
