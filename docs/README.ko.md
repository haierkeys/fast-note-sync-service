[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

문제가 발생하면 [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new)를 생성하거나 텔레그램 그룹에 가입하여 도움을 요청하세요: [https://t.me/obsidian_users](https://t.me/obsidian_users)

중국 본토 사용자의 경우 Tencent cnb.cool 미러 저장소를 사용하는 것을 권장합니다: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


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
  <em>Golang + Websocket + React 기반</em>
</p>

<p align="center">
  데이터 제공은 클라이언트 플러그인과 함께 사용해야 합니다: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## 🎯 주요 기능

* **🧰 MCP (Model Context Protocol) 기본 지원**:
  * `FNS`는 MCP 서버로써 `Cherry Studio`, `Cursor` 등 호환되는 AI 클라이언트에 연결될 수 있어, AI가 개인 노트와 첨부 파일을 읽고 쓸 수 있도록 지원하며 모든 변경 사항은 실시간으로 모든 단말기에 동기화됩니다.
* **🚀 REST API 지원**:
  * 표준 REST API 인터페이스를 제공하여 프로그램 방식(예: 자동화 스크립트, AI 비서 통합)을 통한 Obsidian 노트의 CRUD 작업을 지원합니다.
  * 자세한 내용은 [RESTful API 문서](/docs/REST_API.md) 또는 [OpenAPI 문서](/docs/swagger.yaml)를 참조하세요.
* **💻 웹 관리 패널**:
  * 현대적인 관리 인터페이스가 내장되어 사용자 생성, 플러그인 구성 생성, 저장소 및 노트 콘텐츠 관리를 쉽게 처리할 수 있습니다.
* **🔄 멀티 디바이스 노트 동기화**:
  * Vault (저장소) 자동 생성을 지원합니다.
  * 노트 관리(CRUD)를 지원하며, 변경 사항은 밀리초 단위로 실시간으로 모든 온라인 디바이스에 배포됩니다.
* **🖼️ 첨부 파일 동기화 지원**:
  * 이미지 등 노트 이외 파일의 동기화를 완벽하게 지원합니다.
  * 대용량 첨부 파일의 분할 업로드 및 다운로드를 지원하며, 분할 크기를 구성하여 동기화 효율성을 높일 수 있습니다.
* **⚙️ 구성 동기화**:
  * `.obsidian` 구성 파일 동기화를 지원합니다.
  * `PDF` 읽기 진행 상태 동기화를 지원합니다.
* **📝 노트 이력**:
  * 웹 페이지 또는 플러그인 클라이언트 측에서 각 노트의 수정 이력 버전을 확인할 수 있습니다.
  * (서버 v1.2+ 필요)
* **🗑️ 휴지통**:
  * 노트 삭제 후 자동으로 휴지통으로 이동합니다.
  * 휴지통에서 노트를 복구할 수 있습니다. (첨부 파일 복구 기능은 향후 순차적으로 추가될 예정입니다.)

* **🚫 오프라인 동기화 전략**:
  * 노트 오프라인 편집 시 자동 병합을 지원합니다. (플러그인 설정 필요)
  * 오프라인에서 삭제한 경우 재연결 후 자동으로 보완되거나 동기화 삭제됩니다. (플러그인 설정 필요)

* **🔗 공유 기능**:
  * 노트 공유를 생성하거나 취소할 수 있습니다.
  * 공유된 노트에서 인용된 이미지, 오디오, 비디오 등의 첨부 파일을 자동으로 분석합니다.
  * 공유 방문 통계 기능을 제공합니다.
  * 공유 노트의 액세스 비밀번호를 설정할 수 있습니다.
  * 공유 노트에 대한 단축 링크를 생성할 수 있습니다.
* **📂 폴더 동기화**:
  * 폴더의 생성/이름 변경/이동/삭제 동기화를 지원합니다.

* **🌳 Git 자동화**:
  * 첨부 파일 및 노트에 변경이 발생하면 자동으로 업데이트되어 원격 Git 저장소로 푸시됩니다.
  * 작업 종료 후 시스템 메모리를 자동으로 해제합니다.

* **☁️ 멀티 스토리지 백업 및 단방향 미러 동기화**:
  * S3/OSS/R2/WebDAV/로컬 등 다양한 스토리지 프로토콜을 지원합니다.
  * 전체/증분 ZIP 정기 아카이브 백업을 지원합니다.
  * Vault 리소스의 원격 스토리지 단방향 미러 동기화를 지원합니다.
  * 만료된 백업을 자동으로 정리하고 맞춤 보존 기간을 지원합니다.

* **🗄️ 멀티 데이터베이스 지원**:
  * SQLite, MySQL, PostgreSQL 등 다양한 주요 데이터베이스를 기본적으로 지원하여 개인부터 팀까지 다양한 배포 요구 사항을 충족합니다.

## ☕ 후원 및 지원

- 이 플러그인이 유용하고 지속적인 개발을 지원하고 싶으시다면 아래의 방법으로 후원해 주세요:

  | Ko-fi *중국 외 지역*                                                                             |    | WeChat Pay *중국 국내 지역*                    |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | 또는 | <img src="/docs/images/wxds.png" height="150"> |

  - 후원자 명단:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.ko.md">Support.ko.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.ko.md">Support.ko.md (cnb.cool 미러 저장소)</a>

## ⏱️ 업데이트 로그

- ♨️ [업데이트 로그 보기](/docs/CHANGELOG.ko.md)

## 🗺️ 로드맵 (Roadmap)

- [ ] WebSocket `Protobuf` 전송 형식 지원을 추가하여 동기화 전송 효율을 강화합니다.
- [ ] 기존 권한 부여 메커니즘을 격리 및 최적화하여 전반적인 보안을 향상합니다.
- [ ] Web UI에서 노트 실시간 업데이트를 추가합니다.
- [ ] 클라이언트 간 P2P 메시지 전송을 추가합니다 (노트 및 첨부 파일 제외, LocalSend와 유사한 기능, 클라이언트 측 저장은 지원되지 않으나 서버 측 저장은 가능).
- [ ] 다양한 도움말 문서 보완.
- [ ] 더 많은 인트라넷 통과(릴레이 게이트웨이) 지원.
- [ ] 빠른 배포 계획:
  * 서버 주소(공용 IP)와 계정 자격 증명만 제공하면 FNS 서버 측 배포가 완료됩니다.
- [ ] 기존의 오프라인 노트 병합 방안을 최적화하고 충돌 처리 메커니즘을 추가합니다.

우리는 지속적인 개선을 진행 중이며, 향후 개발 계획은 다음과 같습니다:

> **개선 제안이나 새로운 아이디어가 있으시면 언제든지 issue를 제출하여 공유해 주세요. 적절한 제안은 신중히 평가하여 채택할 것입니다.**

## 🚀 빠른 배포

다양한 설치 방법을 제공합니다. **원클릭 스크립트** 또는 **Docker**를 사용하는 것을 권장합니다.

### 방법 1: 원클릭 스크립트 (권장)

시스템 환경을 자동으로 감지하고 설치 및 서비스 등록을 완료합니다.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

중국 지역은 Tencent `cnb.cool` 미러 소스를 사용할 수 있습니다:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**스크립트의 주요 동작:**

  * 현재 시스템에 적합한 Release 바이너리 파일을 자동으로 다운로드합니다.
  * 기본적으로 `/opt/fast-note`에 설치되며, `/usr/local/bin/fns`에 글로벌 바로 가기 명령 `fns`를 생성합니다.
  * Systemd (Linux) 또는 Launchd (macOS) 서비스를 구성하고 시작하여 자동 실행을 사용 설정합니다.
  * **관리 명령**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **대화형 메뉴**: 직접 `fns`를 실행하여 메뉴로 들어가며, 설치/업그레이드, 서비스 제어, 자동 실행 구성, GitHub과 CNB 미러 간 전환 등을 지원합니다.

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

docker-compose.yaml 파일 생성:

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # RESTful API 및 WebSocket 포트 (/api/user/sync는 WebSocket 인터페이스 주소)
    volumes:
      - ./storage:/fast-note-sync/storage  # 데이터 저장소
      - ./config:/fast-note-sync/config    # 구성 파일
```

서비스 시작:

```bash
docker compose up -d
```

-----

### 방법 3: 수동 바이너리 설치

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases)에서 해당 시스템용 최신 버전을 다운로드하고 압축을 푼 후 실행합니다:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 사용 가이드

1.  **관리 패널 액세스**:
    브라우저에서 `http://{서버 IP}:9000`을 엽니다.
2.  **초기 설정**:
    최초 방문 시 계정을 등록해야 합니다. *(등록 기능을 비활성화하려면 구성 파일에서 `user.register-is-enable: false`로 설정하세요)*
3.  **클라이언트 구성**:
    관리 패널에 로그인하고 **"API 구성 복사"**를 클릭합니다.
4.  **Obsidian에 연결**:
    Obsidian 플러그인 설정 페이지를 열고 방금 복사한 구성 정보를 붙여넣습니다.


## ⚙️ 구성 설명

기본 구성 파일은 `config.yaml`입니다. 프로그램은 **루트 디렉터리** 또는 **config/** 디렉터리에서 자동으로 검색합니다.

전체 구성 예 보기: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx 역방향 프록시 예

전체 구성 예 보기: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) 지원

FNS는 현재 **MCP (Model Context Protocol)**를 기본 지원하며, **SSE** 및 **StreamableHTTP** 두 가지 전송 프로토콜을 모두 제공합니다.

FNS를 MCP 서버로 Cherry Studio, Cursor, Claude Code, hermes-agent 등 호환되는 AI 클라이언트에 직접 연결할 수 있습니다. 연결 후 AI는 개인 노트와 첨부 파일을 읽고 쓸 수 있는 권한을 얻게 됩니다. 또한 MCP를 통한 변경 사항은 WebSocket을 통해 실시간으로 각 클라이언트 디바이스에 동기화됩니다.

### 공통 요청 헤더 매개변수

어떤 전송 모드를 사용하든 상관없이 다음 요청 헤더가 지원됩니다:

- **Authorization Header**: `Authorization: Bearer <사용자 API 토큰>` (WebGUI의 API 구성 복사 옵션에서 가져옴)
- **선택적 Header**: `X-Default-Vault-Name: <노트 보관소 이름>` (도구 호출 시 `vault` 매개변수가 제공되지 않은 경우 MCP 작업의 기본 보관소를 지정)
- **선택적 Header**: `X-Client: <클라이언트 유형>` (MCP에 연결하는 클라이언트 유형, 예: `Cherry Studio`, `OpenClaw`)
- **선택적 Header**: `X-Client-Version: <클라이언트 버전>` (연결하는 클라이언트의 버전, 예: `1.1`)
- **선택적 Header**: `X-Client-Name: <클라이언트 이름>` (연결하는 클라이언트의 이름, 예: `Mac`)

---

### 연결 구성: StreamableHTTP 모드 (권장)

StreamableHTTP는 MCP 에코시스템의 표준 전송 프로토콜입니다. 단일 엔드포인트로 요청을 완료할 수 있어 방화벽 친화적이며 새로운 MCP 클라이언트(예: Claude Code 및 hermes-agent)에서 기본 지원됩니다.

- **인터페이스 주소**: `http://<서버 IP 또는 도메인>:<포트>/api/mcp`
- **요청 메서드**: `POST` (요청/알림 전송), `GET` (서버 푸시 감시), `DELETE` (세션 종료)

#### 예: Claude Code / hermes-agent / Cursor 등

*(참고: `<ServerIP>`, `<Port>`, `<Token>`, `<VaultName>`을 실제 정보로 교체해 주세요)*

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

### 연결 구성: SSE 모드 (하위 호환)

SSE 모드는 레거시 전송 프로토콜입니다. 하위 호환성을 유지하기 위해 완전히 보존되었으며 SSE만 지원하는 MCP 클라이언트(예: Cherry Studio)에 적합합니다.

- **인터페이스 주소**: `http://<서버 IP 또는 도메인>:<포트>/api/mcp/sse`

#### 예: Cherry Studio / Cline 등

*(참고: `<ServerIP>`, `<Port>`, `<Token>`, `<VaultName>`을 실제 정보로 교체해 주세요)*

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

## 🔗 클라이언트 & 클라이언트 플러그인 & 협력 프로젝트

* Obsidian Fast Note Sync 플러그인
  * [Obsidian Fast Note Sync 플러그인](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool 미러 저장소](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* 제3자 클라이언트
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) FNS WebSocket 동기화 프로토콜을 기반으로 Python으로 구현된 실시간 양방향 동기화 명령줄 클라이언트입니다. GUI가 없는 Linux 서버 환경(예: OpenClaw)에 적합하며 Obsidian 데스크톱/모바일 앱과 동등한 동기화 성능을 제공합니다.
  * [go-fast-note-sync](https://github.com/erichll/go-fast-note-sync) FNS WebSocket 동기화 프로토콜을 기반으로 Go로 개발된 Go CLI 백그라운드 동기화 데몬입니다. 주로 Linux 헤드리스 환경을 대상으로 하며 macOS 및 Windows도 지원합니다.
  * [Fast-note-sync-docker](https://github.com/youpingfang/obsidian-note-sync-docker) Docker, Python, FNS WebSocket 동기화 프로토콜을 활용한 빠른 컨테이너화 배포 솔루션으로 노트 저장소와 구성 파일을 원격 서버에 동기화합니다.
* 협력 프로젝트
  * [Share to Save](https://github.com/chenxiccc/Obsidian-Share-to-Save) 공유된 웹페이지 URL을 자동으로 마크다운 노트로 다운로드하는 Obsidian 플러그인입니다.
