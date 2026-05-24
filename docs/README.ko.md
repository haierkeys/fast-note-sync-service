[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

문의 사항이 있으시면 새 [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new)를 생성하시거나, Telegram 커뮤니티 그룹에 가입하여 도움을 요청하세요: [https://t.me/obsidian_users](https://t.me/obsidian_users)

중국 본토 지역에서는 Tencent `cnb.cool` 미러 저장소를 사용하는 것을 권장합니다: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


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
  <em>Golang + WebSocket + React 기반 구축</em>
</p>

<p align="center">
  클라이언트 데이터 제공은 플러그인과 함께 사용해야 합니다: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## 🎯 핵심 기능

* **🧰 MCP (Model Context Protocol) 기본 지원**：
  * `FNS`는 `Cherry Studio`, `Cursor` 등 호환되는 AI 클라이언트에 MCP 서버로 연결하여, AI가 개인 노트 및 첨부 파일을 읽고 쓸 수 있도록 지원하며, 모든 변경 사항은 각 기기로 실시간 동기화됩니다.
* **🚀 REST API 지원**：
  * 표준 REST API 인터페이스를 제공하여 자동화 스크립트, AI 어시스턴트 통합 등 프로그램 방식으로 Obsidian 노트의 생성, 조회, 수정, 삭제(CRUD)를 지원합니다.
  * 자세한 내용은 [RESTful API 문서](/docs/REST_API.md) 또는 [OpenAPI 문서](/docs/swagger.yaml)를 참조하세요.
* **💻 Web 관리 패널**：
  * 현대적인 관리 페이지가 내장되어 사용자 생성, 플러그인 설정 생성, 저장소 및 노트 콘텐츠 관리를 손쉽게 수행할 수 있습니다.
* **🔄 멀티 디바이스 노트 동기화**：
  * **Vault (보관소)** 자동 생성 지원.
  * 노트 관리(추가, 삭제, 수정, 조회)를 지원하며 변경 사항은 밀리초 단위로 실시간으로 모든 온라인 기기에 배포됩니다.
* **🖼️ 첨부 파일 동기화 지원**：
  * 이미지 등 노트가 아닌 파일의 동기화를 완벽하게 지원합니다.
  * 대용량 첨부 파일 분할 업로드 및 다운로드를 지원하고 분할 크기를 설정할 수 있어 동기화 효율이 증가합니다.
* **⚙️ 설정 동기화**：
  * `.obsidian` 설정 파일 동기화를 지원합니다.
  * `PDF` 진행 상태 동기화를 지원합니다.
* **📝 노트 히스토리**：
  * 웹 페이지 또는 플러그인 단에서 각 노트의 과거 수정 버전(히스토리)을 확인할 수 있습니다. (서버 v1.2+ 필요)
* **🗑️ 휴지통**：
  * 노트가 삭제되면 자동으로 휴지통으로 이동합니다.
  * 휴지통에서 노트 복구를 지원합니다. (첨부 파일 복구 기능은 순차적으로 추가될 예정입니다)

* **🚫 오프라인 동기화 정책**：
  * 오프라인 편집 노트의 자동 병합을 지원합니다. (플러그인 설정 필요)
  * 오프라인 삭제 후 재연결 시 자동 복구 또는 삭제 동기화를 지원합니다. (플러그인 설정 필요)

* **🔗 공유 기능**：
  * 노트 공유 설정 및 취소가 가능합니다.
  * 공유된 노트에 포함된 이미지, 오디오, 비디오 등 첨부 파일을 자동으로 분석합니다.
  * 공유 노트의 방문 통계 기능을 제공합니다.
  * 공유 노트의 액세스 비밀번호를 설정할 수 있습니다.
  * 공유 노트의 단축 URL을 생성할 수 있습니다.
* **📂 디렉토리 동기화**：
  * 폴더의 생성, 이름 변경, 이동, 삭제 동기화를 지원합니다.

* **🌳 Git 자동화**：
  * 첨부 파일 및 노트에 변경이 발생하면 자동으로 업데이트하여 원격 Git 저장소에 푸시합니다.
  * 작업 완료 후 시스템 메모리를 자동으로 해제합니다.

* **☁️ 멀티 스토리지 백업 및 단방향 미러 동기화**：
  * S3, OSS, R2, WebDAV, 로컬 등 다양한 스토리지 프로토콜을 지원합니다.
  * 전체/증분 ZIP 정기 아카이브 백업을 지원합니다.
  * Vault 리소스의 원격 스토리지 단방향 미러 동기화를 지원합니다.
  * 만료된 백업 자동 정리 및 보관 기간 설정을 지원합니다.

* **🗄️ 멀티 데이터베이스 지원**：
  * SQLite, MySQL, PostgreSQL 등 다양한 주요 데이터베이스를 네이티브 지원하여 개인부터 팀까지 다양한 배포 요구사항을 만족시킵니다.

## ☕ 후원 및 지원

- 이 플러그인이 유용하고 계속 개발되기를 원하신다면 다음 방법으로 후원해 주세요:

  | Ko-fi *중국 이외 지역*                                                                               |    | 위챗(WeChat) QR 코드 결제 *중국 지역*          |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | 또는 | <img src="/docs/images/wxds.png" height="150"> |

  - 후원자 명단:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md (cnb.cool 미러 저장소)</a>

## ⏱️ 변경 로그

- ♨️ [변경 로그 확인하기](/docs/CHANGELOG.ko.md)

## 🗺️ 로드맵 (Roadmap)

- [ ] WebSocket `Protobuf` 전송 포맷 지원을 추가하여 동기화 전송 효율 강화.
- [ ] 기존 인증 메커니즘을 격리 및 최적화하여 전반적인 보안 향상.
- [ ] WebGui 노트 실시간 업데이트 추가.
- [ ] 클라이언트 간 피어투피어(P2P) 메시지 전송 추가(노트 및 첨부 파일 제외, LocalSend와 유사한 기능. 클라이언트 저장 미지원, 서버 저장 지원).
- [ ] 각종 도움말 문서 보완.
- [ ] 더 많은 인트라넷 침투(릴레이 게이트웨이) 지원.
- [ ] 빠른 배포 계획
  * 서버 주소(공인 IP), 계정, 비밀번호만 제공하면 FNS 서버 배포가 완료됩니다.
- [ ] 기존 오프라인 노트 병합 방안을 최적화하고 충돌 처리 메커니즘 추가.

지속적으로 개선 중이며, 아래는 향후 개발 계획입니다:

> **개선 제안이나 새로운 아이디어가 있으시면 issue를 통해 공유해 주세요. 소중히 검토하여 반영하겠습니다.**

## 🚀 빠른 배포

다양한 설치 방법을 제공하며, **원클릭 스크립트** 또는 **Docker** 사용을 권장합니다.

### 방법 1: 원클릭 스크립트 (권장)

시스템 환경을 자동으로 감지하여 설치 및 서비스를 등록합니다.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

중국 지역에서는 Tencent `cnb.cool` 미러 저장소를 사용할 수 있습니다.
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**스크립트 주요 동작:**

  * 현재 시스템에 맞는 Release 바이너리 파일을 자동으로 다운로드합니다.
  * 기본적으로 `/opt/fast-note`에 설치되며, `/usr/local/bin/fns`에 전역 단축 명령 `fns`를 생성합니다.
  * Systemd(Linux) 또는 Launchd(macOS) 서비스를 설정하고 실행하여 부팅 시 자동 시작을 구현합니다.
  * **관리 명령**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **대화형 메뉴**: `fns`를 직접 실행하면 대화형 메뉴로 진입하여 설치/업그레이드, 서비스 제어, 자동 시작 설정, GitHub / CNB 미러 전환을 수행할 수 있습니다.

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

서비스 실행:

```bash
docker compose up -d
```

-----

### 방법 3: 수동 바이너리 설치

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases)에서 운영체제에 맞는 최신 버전을 다운로드하고 압축을 푼 후 실행합니다:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 사용 가이드

1.  **관리 패널 접속**：
    브라우저에서 `http://{서버 IP}:9000`을 엽니다.
2.  **초기 설정**：
    최초 접속 시 계정을 등록해야 합니다. *(가입 기능을 비활성화하려면 설정 파일에서 `user.register-is-enable: false`로 설정하세요)*
3.  **클라이언트 설정**：
    관리 패널에 로그인한 후 **“API 설정 복사”**를 클릭합니다.
4.  **Obsidian 연결**：
    Obsidian 플러그인 설정 페이지를 열고 방금 복사한 설정 정보를 붙여넣습니다.


## ⚙️ 설정 설명

기본 설정 파일은 `config.yaml`이며, 프로그램은 **루트 디렉토리** 또는 **config/** 디렉토리에서 자동으로 검색합니다.

설정 예시: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx 역방향 프록시 설정 예시

설정 예시: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) 지원

FNS는 **MCP (Model Context Protocol)**를 기본 지원하며, **SSE**와 **StreamableHTTP** 두 가지 전송 프로토콜을 함께 제공합니다.

FNS를 MCP 서버로 사용하여 Cherry Studio, Cursor, Claude Code, hermes-agent 등 호환되는 AI 클라이언트에 직접 연결할 수 있습니다. 연결 후 AI는 개인 노트와 첨부 파일을 읽고 쓸 수 있는 권한을 가지게 됩니다. 동시에 MCP를 통해 생성된 모든 수정 사항은 WebSocket을 통해 각 디바이스에 실시간으로 동기화됩니다.

### 공통 요청 헤더 매개변수

전송 모드에 관계없이 다음 요청 헤더를 지원합니다:

- **인증 헤더**: `Authorization: Bearer <사용자 API 토큰>` (WebGUI의 API 설정 복사에서 획득 가능)
- **선택 헤더**: `X-Default-Vault-Name: <보관소 이름>` (MCP 작업의 기본 보관소를 지정. 도구 호출 시 `vault` 매개변수가 지정되지 않은 경우 이 값이 사용됩니다)
- **선택 헤더**: `X-Client: <클라이언트 유형>` (MCP에 연결하는 클라이언트 유형. 예: Cherry Studio / OpenClaw)
- **선택 헤더**: `X-Client-Version: <클라이언트 버전>` (MCP에 연결하는 클라이언트 버전. 예: 1.1)
- **선택 헤더**: `X-Client-Name: <클라이언트 이름>` (MCP에 연결하는 클라이언트 이름. 예: Mac)

---

### 연결 설정: StreamableHTTP 모드 (권장)

StreamableHTTP는 MCP 생태계의 표준 전송 프로토콜입니다. 단일 엔드포인트에서 요청을 처리하므로 방화벽 친화적이며 최신 MCP 클라이언트(Claude Code, hermes-agent 등)에서 기본 지원합니다.

- **인터페이스 주소**: `http://<서버 IP 또는 도메인>:<포트>/api/mcp`
- **요청 방식**: `POST`(요청/알림 전송), `GET`(서버 푸시 수신), `DELETE`(세션 종료)

#### 예시: Claude Code / hermes-agent / Cursor 등

*(참고: `<ServerIP>`, `<Port>`, `<Token>`, `<VaultName>`을 실제 정보로 변경해 주세요)*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp",
      "type": "http",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer <Token>",
        "X-Default-Vault-Name": "<VaultName>",
        "X-Client": "<Client>",
        "X-Client-Version": "<ClientVersion>",
        "X-Client-Name": "<ClientName>"
      }
    }
  }
}
```

---

### 연결 설정: SSE 모드 (하위 호환)

SSE 모드는 레거시 전송 프로토콜이지만 하위 호환성을 유지하기 위해 그대로 제공되며, SSE만 지원하는 MCP 클라이언트(Cherry Studio 등)에 적합합니다.

- **인터페이스 주소**: `http://<서버 IP 또는 도메인>:<포트>/api/mcp/sse`

#### 예시: Cherry Studio / Cline 등

*(참고: `<ServerIP>`, `<Port>`, `<Token>`, `<VaultName>`을 실제 정보로 변경해 주세요)*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp/sse",
      "type": "sse",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer <Token>",
        "X-Default-Vault-Name": "<VaultName>",
        "X-Client": "<Client>",
        "X-Client-Version": "<ClientVersion>",
        "X-Client-Name": "<ClientName>"
      }
    }
  }
}
```

## 🔗 클라이언트 & 클라이언트 플러그인

* Obsidian Fast Note Sync 플러그인
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool 미러 저장소](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* 서드파티 클라이언트
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) Python과 FNS WebSocket 동기화 프로토콜에 기반해 구현된 실시간 쌍방향 동기화 명령줄 클라이언트. GUI가 없는 Linux 서버 환경(OpenClaw 등)에 적합하며, Obsidian 데스크톱 및 모바일 기기와 대등한 동기화 능력을 지원합니다.
  * [go-fast-note-sync](https://github.com/erichll/go-fast-note-sync) Go와 FNS WebSocket 동기화 프로토콜 기반의 Go CLI 백그라운드 동기화 데몬 프로세스. 주로 헤드리스(headless) Linux 환경을 대상으로 하며, macOS와 Windows도 지원합니다.
  * [Fast-note-sync-docker](https://github.com/youpingfang/obsidian-note-sync-docker) Docker, Python, FNS WebSocket 동기화 프로토콜에 기반한 신속한 컨테이너화 배포 방안. 노트 보관소와 설정 파일을 원격 서버로 동기화합니다.
