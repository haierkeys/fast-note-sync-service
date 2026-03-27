[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

문제가 있으시면 새 [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new)를 등록하거나, Telegram 그룹에 참여해 도움을 받으세요: [https://t.me/obsidian_users](https://t.me/obsidian_users)

중국 본토 사용자에게는 Tencent `cnb.cool` 미러 사용을 권장합니다: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)



<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>고성능, 저지연 노트 동기화, 온라인 관리 및 원격 REST API 서비스 플랫폼</strong>
  <br>
  <em>Golang + WebSocket + SQLite + React 기반으로 구축</em>
</p>

<p align="center">
  데이터 제공을 위해 클라이언트 플러그인이 필요합니다: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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
    * 표준 REST API 인터페이스를 제공하여 프로그래밍 방식(자동화 스크립트, AI 어시스턴트 연동 등)으로 Obsidian 노트의 CRUD 작업을 지원합니다.
    * 자세한 내용은 [RESTful API 문서](/docs/REST_API.md) 또는 [OpenAPI 문서](/docs/swagger.yaml)를 참고하세요.
* **💻 웹 관리 패널**:
  * 내장된 현대적 관리 인터페이스로 사용자 생성, 플러그인 설정 생성, Vault 및 노트 콘텐츠 관리를 쉽게 할 수 있습니다.
* **🔄 멀티 디바이스 노트 동기화**:
    * **Vault(저장소)** 자동 생성을 지원합니다.
    * 노트 관리(추가, 삭제, 수정, 조회)를 지원하며, 변경 사항은 밀리초 단위로 모든 온라인 기기에 실시간 배포됩니다.
* **🖼️ 첨부파일 동기화 지원**:
    * 이미지 등 노트 외 파일의 동기화를 완벽하게 지원합니다.
    * 대용량 첨부파일의 청크 업로드/다운로드를 지원하며, 청크 크기를 설정하여 동기화 효율을 높일 수 있습니다.
* **⚙️ 설정 동기화**:
    * `.obsidian` 설정 파일의 동기화를 지원합니다.
    * `PDF` 진행 상태 동기화를 지원합니다.
* **📝 노트 히스토리**:
    * 웹 페이지 및 플러그인 측에서 각 노트의 이전 수정 버전을 확인할 수 있습니다.
    * (서버 v1.2+ 필요)
* **🗑️ 휴지통**:
    * 削除된 노트가 자동으로 휴지통으로 이동합니다.
    * 휴지통에서 노트를 복원할 수 있습니다. (첨부파일 복원 기능은 이후 순차적으로 추가 예정)

* **🚫 오프라인 동기화 전략**:
    * 오프라인으로 편집한 노트의 자동 병합을 지원합니다. (플러그인 측 설정 필요)
    * 오프라인 삭제는 재연결 후 자동으로 보완 또는 동기화됩니다. (플러그인 측 설정 필요)

* **🔗 공유 기능**:
    * 노트 공유를 생성/취소할 수 있습니다.
    * 공유 노트에서 참조된 이미지, 오디오, 비디오 등 첨부파일을 자동으로 분석합니다.
    * 공유 접속 통계 기능을 제공합니다.
* **📂 디렉토리 동기화**:
    * 폴더의 생성/이름 변경/이동/삭제 동기화를 지원합니다.

* **🌳 Git 자동화**:
    * 첨부파일이나 노트에 변경이 생기면 원격 Git 저장소에 자동으로 업데이트 및 푸시합니다.
    * 작업 완료 후 시스템 메모리를 자동으로 해제합니다.

* **☁️ 다중 스토리지 백업 및 단방향 미러 동기화**:
    * S3/OSS/R2/WebDAV/로컬 등 다양한 스토리지 프로토콜을 지원합니다.
    * 전체/증분 ZIP 정기 아카이브 백업을 지원합니다.
    * Vault 리소스를 원격 스토리지로 단방향 미러 동기화를 지원합니다.
    * 만료된 백업 자동 정리, 보존 일수를 사용자 정의할 수 있습니다.

## ☕ 후원 및 지원

- 이 프로젝트가 유용하고 지속적인 개발을 지원하고 싶으시다면 다음 방법으로 지원해 주세요:

  | Ko-fi *(중국 외 지역)*                                                                            |    | WeChat Pay *(중국)*                            |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | 또는 | <img src="/docs/images/wxds.png" height="150"> |

  - 후원자 목록:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md (cnb.cool 미러)</a>

## ⏱️ 변경 이력

- ♨️ [변경 이력 보기](/docs/CHANGELOG.ko.md)

## 🗺️ 로드맵 (Roadmap)

지속적으로 개선 중입니다. 다음은 향후 개발 계획입니다:


- [x] **🤖 MCP 지원**: AI MCP 관련 기능 지원 추가.
- [ ] **더 많은 데이터베이스 유형 지원**


> **개선 제안이나 새로운 아이디어가 있으시면 issue를 통해 자유롭게 공유해 주세요. 적합한 제안은 신중하게 검토하여 반영하겠습니다.**

## 🚀 빠른 배포

여러 설치 방법을 제공합니다. **원클릭 스크립트** 또는 **Docker** 사용을 권장합니다.

### 방법 1: 원클릭 스크립트 (권장)

시스템 환경을 자동으로 감지하여 설치 및 서비스 등록을 완료합니다.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

중국 사용자는 Tencent `cnb.cool` 미러를 사용할 수 있습니다:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**스크립트 주요 동작:**

  * 현재 시스템에 맞는 Release 바이너리를 자동 다운로드합니다.
  * 기본적으로 `/opt/fast-note`에 설치하며, `/usr/local/bin/fns`에 전역 단축 명령어 `fns`를 생성합니다.
  * Systemd(Linux) 또는 Launchd(macOS) 서비스를 설정하고 시작하여 부팅 시 자동 시작합니다.
  * **관리 명령어**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **대화형 메뉴**: `fns`를 직접 실행하면 대화형 메뉴가 열려 설치/업그레이드, 서비스 제어, 부팅 자동 시작 설정, GitHub/CNB 미러 간 전환을 지원합니다.

-----

### 방법 2: Docker 배포

#### Docker Run

```bash
# 1. 이미지 풀
docker pull haierkeys/fast-note-sync-service:latest

# 2. 컨테이너 시작
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 \
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
      - "9000:9000"  # RESTful API & WebSocket 포트; /api/user/sync 가 WebSocket 엔드포인트
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

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases)에서 시스템에 맞는 최신 버전을 다운로드하고 압축 해제 후 실행:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 사용 가이드

1.  **관리 패널 접속**:
    브라우저에서 `http://{서버IP}:9000`을 엽니다.
2.  **초기 설정**:
    최초 접속 시 계정을 등록합니다. *(등록 기능을 비활성화하려면 설정 파일에서 `user.register-is-enable: false`로 설정하세요)*
3.  **클라이언트 설정**:
    관리 패널에 로그인하고 **"API 설정 복사"** 를 클릭합니다.
4.  **Obsidian 연결**:
    Obsidian 플러그인 설정 페이지를 열고 복사한 설정 정보를 붙여넣습니다.


## ⚙️ 설정 설명

기본 설정 파일은 `config.yaml`이며, 프로그램은 **루트 디렉토리** 또는 **config/** 디렉토리에서 자동으로 찾습니다.

전체 설정 예시 보기: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx 리버스 프록시 설정 예시

전체 설정 예시 보기: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (모델 컨텍스트 프로토콜) 지원

FNS는 이제 **MCP (Model Context Protocol)** 를 네이티브로 지원합니다.

FNS를 MCP 서버로서 Cherry Studio, Cursor 등의 호환 AI 클라이언트에 직접 연결할 수 있습니다. 연결 후 AI는 개인 노트와 첨부파일을 읽고 쓸 수 있게 됩니다. 또한 MCP에 의해 발생한 모든 변경 사항은 WebSocket을 통해 각 기기 단말기에 실시간으로 동기화됩니다.

### 연결 설정 (SSE 모드)

FNS는 **SSE 프로토콜**을 통해 MCP 인터페이스를 제공합니다. 일반 파라미터는 다음과 같습니다:
- **엔드포인트 URL**: `http://<서버 IP 또는 도메인>:<포트>/api/mcp/sse`
- **인증 헤더**: `token: <API 토큰>` (WebGUI의 "API 설정 복사"에서 획득)


#### 예시: Cherry Studio

Cherry Studio의 MCP Server 설정에 다음 JSON 노드를 추가하세요.
*（참고: `<ServerIP>`, `<Port>`, `<Token>`을 실제 정보로 교체하세요）*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp/sse",
      "type": "sse",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer <Token>"
      }
    }
  }
}
```

## 🔗 클라이언트 & 클라이언트 플러그인

* Obsidian Fast Note Sync 플러그인
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool 미러](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* 서드파티 클라이언트
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) — Python과 FNS WS 인터페이스를 기반으로 한 양방향 실시간 동기화 커맨드라인 클라이언트. GUI가 없는 Linux 서버 환경(OpenClaw 등)에 적합하며, Obsidian 데스크탑/모바일과 동등한 동기화 능력을 구현합니다.
