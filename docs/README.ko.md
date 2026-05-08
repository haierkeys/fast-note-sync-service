[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

문제가 발생하면 새 [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new)를 생성하거나, 텔레그램 그룹에 참여하여 도움을 받으세요: [https://t.me/obsidian_users](https://t.me/obsidian_users)

중국 본토 사용자에게는 Tencent `cnb.cool` 미러 사용을 권장합니다: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>고성능·저지연 노트 동기화, 온라인 관리, 원격 REST API 서비스 플랫폼</strong>
  <br>
  <em>Golang + WebSocket + React 기반으로 구축</em>
</p>

<p align="center">
  데이터 동기화에는 클라이언트 플러그인이 필요합니다：<a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

* **🧰 MCP (Model Context Protocol) 네이티브 지원**：
  * `FNS`는 MCP 서버로서 `Cherry Studio`, `Cursor` 등 호환 AI 클라이언트에 연결되어, AI가 개인 노트와 첨부 파일을 읽고 쓸 수 있으며, 모든 변경 사항은 모든 기기에 실시간으로 동기화됩니다.
* **🚀 REST API 지원**：
  * 표준 REST API 엔드포인트를 제공하여 프로그래밍 방식(자동화 스크립트, AI 어시스턴트 통합 등)으로 Obsidian 노트의 CRUD 작업을 지원합니다.
  * 자세한 내용은 [RESTful API 문서](/docs/REST_API.md) 또는 [OpenAPI 문서](/docs/swagger.yaml)를 참조하세요.
* **💻 웹 관리 패널**：
  * 현대적인 내장 관리 인터페이스로 사용자 생성, 플러그인 설정 생성, Vault 및 노트 콘텐츠 관리를 손쉽게 수행할 수 있습니다.
* **🔄 멀티 디바이스 노트 동기화**：
  * **Vault(저장소)** 자동 생성 지원.
  * 노트 관리(생성, 삭제, 수정, 검색)를 지원하며 변경 사항을 밀리초 단위로 모든 온라인 기기에 실시간 배포합니다.
* **🖼️ 첨부 파일 동기화 지원**：
  * 이미지 등 노트 이외의 파일 동기화를 완벽 지원.
  * 대용량 첨부 파일의 청크 업로드/다운로드를 지원하며 청크 크기 설정이 가능하여 동기화 효율을 향상시킵니다.
* **⚙️ 설정 동기화**：
  * `.obsidian` 설정 파일 동기화 지원.
  * `PDF` 읽기 진행 상태 동기화 지원.
* **📝 노트 이력**：
  * 웹 패널 및 플러그인 측에서 각 노트의 과거 수정 버전을 확인할 수 있습니다.
  * (서버 v1.2+ 필요)
* **🗑️ 휴지통**：
  * 삭제된 노트는 자동으로 휴지통으로 이동합니다.
  * 휴지통에서 노트 복원 지원. (첨부 파일 복원 기능은 향후 순차적으로 추가 예정)

* **🚫 오프라인 동기화 전략**：
  * 오프라인 편집 노트의 자동 병합 지원. (플러그인 측 설정 필요)
  * 오프라인 삭제는 재연결 후 자동으로 보완 또는 삭제 동기화됩니다. (플러그인 측 설정 필요)

* **🔗 공유 기능**：
  * 노트 공유 링크 생성/취소 가능.
  * 공유 노트에서 참조된 이미지, 오디오, 비디오 등 첨부 파일 자동 파싱.
  * 공유 접근 통계 기능 제공.
  * 공유 노트 접근 비밀번호 설정 지원.
  * 공유 노트 단축 링크 생성 지원.
* **📂 디렉토리 동기화**：
  * 폴더의 생성/이름 변경/이동/삭제 동기화 지원.

* **🌳 Git 자동화**：
  * 첨부 파일이나 노트가 변경되면 자동으로 원격 Git 저장소에 커밋 및 푸시합니다.
  * 작업 완료 후 시스템 메모리를 자동으로 해제합니다.

* **☁️ 멀티 스토리지 백업 & 단방향 미러 동기화**：
  * S3, OSS, R2, WebDAV, 로컬 등 다양한 스토리지 프로토콜 지원.
  * 전체/증분 ZIP 정기 아카이브 백업 지원.
  * Vault 리소스의 원격 스토리지 단방향 미러 동기화 지원.
  * 만료된 백업 자동 정리, 보존 일수 커스터마이징 지원.

* **🗄️ 멀티 데이터베이스 지원**：
  * SQLite, MySQL, PostgreSQL 등 주요 데이터베이스를 네이티브로 지원하여 개인부터 팀까지 다양한 배포 요구를 충족합니다.

## ☕ 후원 및 지원

- 이 프로젝트가 유용하고 지속적인 개발을 지원하고 싶다면, 다음 방법으로 지원해 주세요:

  | Ko-fi *（중국 본토 이외）*                                                                               |    | WeChat 기부 *（중국 본토）*                            |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | 또는 | <img src="/docs/images/wxds.png" height="150"> |

  - 후원자 목록:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md (cnb.cool 미러)</a>

## ⏱️ 업데이트 로그

- ♨️ [업데이트 로그 보기](/docs/CHANGELOG.ko.md)

## 🗺️ 로드
- [ ] WebSocket `Protobuf` 전송 형식 지원 추가로 동기화 전송 효율 강화.
- [ ] 기존 인증 메커니즘 분리 및 최적화로 전체 보안성 향상.
- [ ] WebGui 노트 실시간 업데이트 추가.
- [ ] 클라이언트 간 P2P 메시지 전송 추가(노트 & 첨부 파일 제외, localsend 유사 기능, 클라이언트 저장 불가, 서버 저장 가능).
- [ ] 각종 도움말 문서 보완.
- [ ] 더 많은 내부 네트워크 관통(릴레이 게이트웨이) 지원.
- [ ] 빠른 배포 계획:
  * 서버 주소(공개), 계정, 비밀번호만 제공하면 FNS 서버 배포 완료.
- [ ] 기존 오프라인 노트 병합 방식 최적화 및 충돌 처리 메커니즘 추가.

지속적으로 개선 중입니다. 다음은 향후 개발 계획입니다:

> **개선 제안이나 새로운 아이디어가 있다면, issue를 제출하여 공유해 주세요——적합한 제안을 진지하게 평가하고 채택하겠습니다.**

## 🚀 빠른 배포

여러 가지 설치 방법을 제공합니다. **원클릭 스크립트** 또는 **Docker**를 권장합니다.

### 방법 1：원클릭 스크립트（권장）

시스템 환경을 자동 감지하고 설치 및 서비스 등록을 완료합니다.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

중국 본토 사용자는 Tencent `cnb.cool` 미러를 사용할 수 있습니다:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**스크립트 주요 동작：**

  * 현재 시스템에 맞는 Release 바이너리를 자동 다운로드.
  * 기본적으로 `/opt/fast-note`에 설치하고 `/usr/local/bin/fns`에 전역 단축 명령어 `fns`를 생성.
  * Systemd(Linux) 또는 Launchd(macOS) 서비스를 설정 및 시작하여 부팅 시 자동 시작 구현.
  * **관리 명령어**：`fns [install|uninstall|start|stop|status|update|menu]`
  * **인터랙티브 메뉴**：`fns`를 직접 실행하면 인터랙티브 메뉴에 진입하며, 설치/업그레이드, 서비스 제어, 자동 시작 설정, GitHub / CNB 미러 전환을 지원.

-----

### 방법 2：Docker 배포

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

`docker-compose.yaml` 파일 생성：

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # RESTful API & WebSocket 포트；/api/user/sync는 WebSocket 엔드포인트
    volumes:
      - ./storage:/fast-note-sync/storage  # 데이터 저장소
      - ./config:/fast-note-sync/config    # 설정 파일
```

서비스 시작：

```bash
docker compose up -d
```

-----

### 방법 3：수동 바이너리 설치

[Releases](https://github.com/haierkeys/fast-note-sync-service/releases)에서 해당 OS의 최신 버전을 다운로드하고 압축 해제 후 실행：

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 사용 가이드

1.  **관리 패널 접근**：
    브라우저에서 `http://{서버IP}:9000`을 엽니다.
2.  **초기 설정**：
    첫 접근 시 계정을 등록합니다. *(등록 기능을 비활성화하려면 설정 파일에서 `user.register-is-enable: false`를 설정하세요)*
3.  **클라이언트 설정**：
    관리 패널에 로그인하고 **"API 설정 복사"**를 클릭합니다.
4.  **Obsidian 연결**：
    Obsidian 플러그인 설정 페이지를 열고 복사한 설정 정보를 붙여넣습니다.


## ⚙️ 설정 안내

기본 설정 파일은 `config.yaml`이며, 프로그램은 **루트 디렉토리** 또는 **config/** 디렉토리에서 자동으로 검색합니다.

전체 설정 예시 확인：[config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx 리버스 프록시 설정 예시

전체 설정 예시 확인：[https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) 지원

FNS는 이제 **MCP (Model Context Protocol)**를 네이티브로 지원하며, **SSE** 및 **StreamableHTTP** 두 가지 전송 프로토콜을 모두 제공합니다.

FNS를 MCP 서버로서 Cherry Studio, Cursor, Claude Code, hermes-agent 등 호환 AI 클라이언트에 직접 연결할 수 있습니다. 연결 후 AI는 개인 노트와 첨부 파일을 읽고 쓸 수 있습니다. MCP에 의해 생성된 모든 변경 사항은 WebSocket을 통해 모든 기기에 실시간으로 동기화됩니다.

### 공통 요청 헤더

전송 모드에 관계없이 다음 헤더가 지원됩니다：

- **인증 헤더**：`Authorization: Bearer <API Token>` （WebGUI의 API 설정 복사에서 획득）
- **선택적 헤더**：`X-Default-Vault-Name: <Vault명>` （MCP 작업의 기본 Vault 지정；도구 호출 시 `vault` 파라미터가 지정되지 않으면 이 값 사용）
- **선택적 헤더**：`X-Client: <클라이언트 타입>` （MCP 연결 클라이언트 타입, 예：Cherry Studio / OpenClaw）
- **선택적 헤더**：`X-Client-Version: <클라이언트 버전>` （MCP 연결 클라이언트 버전, 예：1.1）
- **선택적 헤더**：`X-Client-Name: <클라이언트 이름>` （MCP 연결 클라이언트 이름, 예：Mac）

---

### 연결 설정：StreamableHTTP 모드（권장）

StreamableHTTP는 MCP 생태계의 표준 전송 프로토콜입니다. 단일 엔드포인트로 모든 요청을 처리하며, 방화벽에 더 친화적이고 새로운 MCP 클라이언트(Claude Code, hermes-agent 등)에서 네이티브로 지원됩니다.

- **엔드포인트**：`http://<서버IP 또는 도메인>:<포트>/api/mcp`
- **메서드**：`POST`（요청/알림 전송）, `GET`（서버 전송 이벤트 수신）, `DELETE`（세션 종료）

#### 예시：Claude Code / hermes-agent / Cursor 등

*（참고：`<ServerIP>`, `<Port>`, `<Token>`, `<VaultName>`을 실제 정보로 교체하세요）*

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

### 연결 설정：SSE 모드（하위 호환）

SSE 모드는 레거시 전송 프로토콜로, 하위 호환성을 위해 완전히 유지됩니다. SSE만 지원하는 MCP 클라이언트(Cherry Studio 등)에 적합합니다.

- **엔드포인트**：`http://<서버IP 또는 도메인>:<포트>/api/mcp/sse`

#### 예시：Cherry Studio / Cline 등

*（참고：`<ServerIP>`, `<Port>`, `<Token>`, `<VaultName>`을 실제 정보로 교체하세요）*

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
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool 미러](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* 서드파티 클라이언트
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) — Python 기반으로 FNS WebSocket API를 통해 양방향 실시간 동기화를 구현한 커맨드라인 클라이언트. GUI가 없는 Linux 서버 환경(OpenClaw 등)을 위해 설계되었으며, Obsidian 데스크톱/모바일 클라이언트와 동등한 동기화 능력을 제공합니다.
