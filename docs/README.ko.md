[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

문제가 발생한 경우, 새로운 [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new)를 생성하시거나 Telegram 커뮤니티 그룹에 가입하여 도움을 요청해 주세요: [https://t.me/obsidian_users](https://t.me/obsidian_users)

중국 본토 지역의 경우, Tencent `cnb.cool`의 미러 리포지토리 사용을 권장합니다: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>고성능, 저지연의 노트 동기화, 온라인 관리, 원격 REST API 서비스 플랫폼</strong>
  <br>
  <em>Golang + Websocket + React 기반으로 구축</em>
</p>

<p align="center">
  데이터를 사용하려면 클라이언트 플러그인과 함께 사용해야 합니다: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

* **🧰 MCP (Model Context Protocol) 기본 지원**:
  * `FNS`는 MCP 서버로서 `Cherry Studio`, `Cursor`와 같은 호환 가능한 AI 클라이언트에 연결할 수 있습니다. 이를 통해 AI가 사용자 개인 노트 및 첨부 파일을 읽고 쓸 수 있는 기능을 갖게 되며, 모든 변경 사항은 실시간으로 각 클라이언트 디바이스와 동기화됩니다.
* **🚀 REST API 지원**:
  * 표준 REST API 인터페이스를 제공하며, 자동화 스크립트나 AI 어시스턴트 통합 같은 프로그래밍 방식으로 Obsidian 노트를 생성, 읽기, 수정 및 삭제(CRUD)할 수 있도록 지원합니다.
  * 자세한 내용은 [RESTful API 문서](/docs/REST_API.md) 또는 [OpenAPI 문서](/docs/swagger.yaml)를 참조해 주세요.
* **💻 웹 관리 패널**:
  * 최신 관리 인터페이스가 내장되어 있어, 사용자 생성, 플러그인 구성 생성, Vault(저장소) 및 노트 내용을 쉽게 관리할 수 있습니다.
* **🔄 다중 장치 노트 동기화**:
  * **Vault (저장소)** 자동 생성을 지원합니다.
  * 노트 관리(추가, 삭제, 수정, 조회)를 지원하며, 변경 사항은 밀리초 단위로 온라인 상태의 모든 장치에 실시간 분산 전송됩니다.
* **🖼️ 첨부 파일 동기화 지원**:
  * 이미지 등 노트 외 파일의 동기화를 완벽하게 지원합니다.
  * 큰 첨부 파일의 분할 업로드/다운로드를 지원하며 분할 크기도 설정할 수 있으므로, 동기화 효율을 향상시켰습니다.
* **⚙️ 설정 동기화**:
  * `.obsidian` 설정 파일의 동기화를 지원합니다.
  * `PDF` 진행 상태 동기화를 지원합니다.
* **📝 노트 내역 기능**:
  * 웹 관리 화면과 클라이언트 플러그인 애플리케이션 내에서 각 노트의 과거 수정 버전을 확인할 수 있습니다.
  * (서버 v1.2+ 버전 필요)
* **🗑️ 휴지통 기능**:
  * 노트를 삭제한 후 이가 휴지통에 자동으로 들어가는 것을 지원합니다.
  * 휴지통에서 노트를 복구하는 기능을 지원합니다. (첨부 파일의 복구 기능도 지속적으로 추가될 예정입니다)

* **🚫 오프라인 동기화 모드 전략**:
  * 오프라인 상태에서 노트 편집에 대해 자동으로 병합하는 것을 지원합니다. (플러그인 옵션 설정이 필요합니다)
  * 오프라인에서 항목을 삭제한 경우, 재연결 후 서버측에서 자동으로 삭제된 노트를 보완하거나 삭제된 로컬과 동기화합니다. (플러그인 설정 필요)

* **🔗 공유 기능**:
  * 노트의 '공유'에 대한 생성 및 취소가 가능합니다.
  * 공유 노트에 삽입된 이미지, 오디오, 비디오 등 첨부 파일을 자동으로 분석합니다.
  * 공유 노트 접속 통계 기능을 제공합니다.
  * 공유 노트에 접속 암호를 설정할 수 있습니다.
  * 공유된 노트에 대한 짧은 링크 생성을 지원합니다.
* **📂 디렉토리(폴더) 동기화**:
  * 폴더의 생성, 이름 변경, 이동, 삭제 항목의 동기화를 지원합니다.

* **🌳 Git 자동화**:
  * 노트 및 파일에 변경이 있을 때 자동으로 Git 원격 저장소로 내역을 업데이트하고 푸시 작업을 수행합니다.
  * 모든 동기화 작업이 종료된 후, 메모리를 반납하여 시스템 활용성을 높입니다.

* **☁️ 다중 스토리지 백업 및 단방향 미러 동기화**:
  * S3, OSS, R2, WebDAV, 로컬 볼륨 등의 다양한 스토리지 프로토콜과 호환됩니다.
  * 전체 / 증분 ZIP 예약 보관 데이터 백업을 수행할 수 있습니다.
  * Vault 리소스를 원격 스토리지에 대해 단방향 미러 동기화가 가능합니다.
  * 기간이 만료된 백업을 자동으로 정리하고, 사용자가 필요한 보존 날짜(유지 보관일수) 지정을 지원합니다.

* **🗄️ 다중 데이터베이스 지원**:
  * SQLite, MySQL, PostgreSQL 등 주류 데이터베이스를 네이티브로 지원하여, 개인 단위부터 팀 협업 모델까지 필요한 배포 요구를 수용합니다.

## ☕ 후원 및 지원

- 이 플러그인이 유용하며 개발이 지속되기를 원하신다면 다음 채널을 통해 후원 부탁드립니다:

  | Ko-fi *중국 이외 지역*                                                                               |    | WeChat QR 기부 *중국 지역*                        |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | 또는 | <img src="/docs/images/wxds.png" height="150"> |

  - 스폰서 목록:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.ko.md">Support.ko.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.ko.md">Support.ko.md (cnb.cool 미러)</a>

## ⏱️ 업데이트 로그 (Changelog)

- ♨️ [업데이트 로그 확인하기](/docs/CHANGELOG.ko.md)

## 🗺️ 로드맵 (Roadmap)

- [ ] 각 레이어에 대대적인 **Mock** 테스트 추가.
- [ ] WebSocket 계층에서 `Protobuf` 통신 프로토콜을 구현하여 동기화 전송 효율을 향상.
- [ ] 백엔드 단에서 동기화 로그 및 동작 로그 등, 시스템 및 유저 감사/운영 기록에 대해 조회가 가능한 기능 생성.
- [ ] 전체 보안을 최상위 상태로 끌어올리도록, 현행 인증 메커니즘을 격리하고 최적화를 달성.
- [ ] WebGui에서도 노트 업데이트 내역을 실시간으로 반영 가능.
- [ ] 클라이언트 간 P2P (Peer-to-Peer) 메시지 전송 기능 추가 (노트 & 첨부파일이 아님, localsend 방식. 클라이언트 간의 저장 없이 서버에 기록 보전됨)
- [ ] 각 분야의 도움말 / 사용 설명서 개선.
- [ ] 인트라넷 환경에서의 우회 액세스 기능 강화 (포트 포워딩 릴레이 지원 등 포함).
- [ ] 빠른 시스템 구축 및 배포 계획:
  * 인프라 서버의 IP(공용 IP)와 계정 패스워드만을 공급하면 단숨에 FNS 서버를 런칭하고 운용.
- [ ] 현존하는 오프라인 노트 병합 전략을 대폭 개선하고 충돌 처리 알고리즘 추가 반영.

우리는 지속적으로 개선 작업을 진행 중이며, 앞으로의 개발 계획은 다음과 같습니다:

> **개선 아이디어나 새로운 제안이 있을 때 Issue를 남겨서 우리에게 의견을 남겨주시면 감사드리겠습니다. 우리는 제출받은 각 항목에 대해 신중히 평가하여 적합한 의견을 도용할 것입니다.**

## 🚀 빠른 배포 가이드

우리는 사용자의 설치 편의성을 고려하여 다양한 옵션을 수록하였습니다.가장 권장하는 방식은 **원클릭 스크립트**와 **Docker** 설치 방식입니다.

### 첫 번째: 원클릭 스크립트 (권장)

시스템 환경 요소의 사양을 자체적으로 검색하여 필수 프로그램을 설치 및 최적의 서비스 등록을 구성합니다.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

중국 대륙 내부에서는 Tencent `cnb.cool`의 미러 다운로드 권한을 사용하십시오:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**스크립트 동작 가이드 프로세스:**

  * 가장 최신의 설치 환경 내장 Release 바이너리 응용프로그램을 다운로드합니다.
  * 기본 디폴트 경로(/opt/fast-note)로 파일들이 압축 해제되며, 시스템 환경변수에 단축 명령인 `fns`(/usr/local/bin/fns)를 저장합니다.
  * Systemd (Linux 기반) 및 Launchd (macOS 환경) 등에 알맞게 서비스가 환경에 저장되어 부수적인 절차 없이 운영 체제가 런칭할 때 함께 자사 프로그램이 동작합니다.
  * **전역 애플리케이션 명령어**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **대화식 콘솔창 지원 기능**: 구동 중인 앱에다 추가 파라미터가 배제된 `fns` 명령을 입력하게 되면 대화형의 화면 기반 콘솔이 뜹니다. 설치, 업데이트/앱 제어/자동 실행 컨테이너 적용 외 GitHub 및 CNB 클라우드 리포지토리 미러 간 변경이 손쉽게 작동 가능합니다.

-----

### 두 번째: Docker 배포 구성방식

#### Docker Run

```bash
# 1. 컨테이너 이미지 불러오기
docker pull haierkeys/fast-note-sync-service:latest

# 2. 이미지 컨테이너 기동시키기
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

로컬 컴퓨터에 `docker-compose.yaml` 파일을 하나 구축합니다：

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # RESTful API & WebSocket 통신 포트 / 내부 설정상 /api/user/sync를 웹소켓 파이프라인으로 운용합니다
    volumes:
      - ./storage:/fast-note-sync/storage  # 데이터/볼륨 스토리지 연동영역
      - ./config:/fast-note-sync/config    # 클라우드/컨테이너 시스템을 위한 Config 마운트
```

서비스 작동:

```bash
docker compose up -d
```

-----

### 세 번째: 수동 바이너리 설치

본인의 운영 체제 버전과 일치하는 팩을 [Releases](https://github.com/haierkeys/fast-note-sync-service/releases)에서 다운로드 받은 후 적절한 디렉토리에서 스크립트를 기동하십시오:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 사용 설명서

1.  **관리자 화면 접속**:
    각자의 브라우저 주소창에 `http://{운용 중인 서버 IP 주소}:9000` 로컬 네트워크 연결.
2.  **프로젝트 가동 세팅 구성**:
    최초 접속 시 계정 생성이 필수로 적용됩니다. *(혹시 외부 가입 통제를 기획 중인 관리자는 서버 Config 파일에 `user.register-is-enable: false` 입력 권장)*
3.  **클라이언트와의 API 조인 연동법**:
    서버 관리 콘솔/로그인 후 우측의 **“API 복사 (Copy API Configuration)”** 버튼 누름.
4.  **Obsidian 옵션 조립 및 세팅**:
    Obsidian 플러그인 설정 창의 셋업 페이지로 향한 후 방금 복사해 둔 환경 데이터를 그대로 붙여 넣기.


## ⚙️ 설정 파일 정보

서버 엔진의 기본 파일 양식 설정 명칭은 « `config.yaml` »이며, 이 프로그램은 환경 디렉터리를 가동 중인 **메인 Root 공간**이거나 자체 **config/**라는 환경 설정 디렉터리 내에서 자동 스캐닝하게 됩니다.

Configuration 샘플 코드를 점검 및 학습하십시오: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx 리버스 프록시 적용 예제

리버스 프록시(HTTPS 접속 용도) 전체 적용 사례 열람 안내: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) 지원

FNS는 현재 **MCP (Model Context Protocol)**를 기본적으로 지원합니다.

FNS를 MCP 서버로 설치하여 Cherry Studio, Cursor 등과 같이 호환되는 AI 클라이언트에 직접 접속하게 연동시킬 수 있습니다. 연동 시점 직후, AI 개체 및 서비스 엔진은 사유 데이터(개인 비메모 및 관련 어태치먼트)의 읽기, 편집 역할을 갖게 됩니다. 더불어 MCP로부터 송부되는 일련의 작업 결과들은 웹 소켓을 경유해 사용 중인 기타의 모든 장치 환경에 실시간 주입 및 저장 단계를 밟습니다.

### 접속 구성 안내 (SSE 모드)

FNS 엔진은 다용도 서버로서 MCP 호출 및 기능을 원격지의 사용자에게 제공할수 있게 **SSE 프로토콜**을 할당합니다. 접속 통신의 요구 파라미터는 다음과 같이 지정되었습니다:
- **인터페이스 주소(URL)**: `http://<호스트 서버의 IP 주소 및 도메인 주소명>:<포트>/api/mcp/sse`
- **강화 인증 (Header)**: `Authorization: Bearer <WebGUI 패널 내부 API 복사 영역에서 배급받은 각자의 API 토큰>`
- **추가 옵션 (Header)**: `X-Default-Vault-Name: <선택한 노트 Vault 명칭>` (MCP 동작 시의 기본 타깃 Vault 구성 목적 / 호출 시점에서 `vault` 변수가 공백일 케이스에 이 구성값이 대체적용)
- **추가 옵션 (Header)**: `X-Client: <클라이언트 종류>` (MCP 연결에 활용하는 구체적인 클라이언트 유형. 예: Cherry Studio / OpenClaw)
- **추가 옵션 (Header)**: `X-Client-Version: <클라이언트 버전>` (MCP에 연결되는 클라이언트 서비스 종류의 버전값. 예: 1.1)
- **추가 옵션 (Header)**: `X-Client-Name: <클라이언트 이름>` (서버 연결을 기동시키는 플랫폼/클라이언트의 운영 명칭. 예: Mac)



#### 예시: Cherry Studio / Cursor / Cline 등 세팅 가이드

운용하시려는 MCP 도메인 서비스 측의 애플리케이션 접속 클라이언트 측에 하단의 설정 코드를 병기 및 수정 후 추가 기입할 것을 권장합니다:
*(안내: 파라미터로 지정된 코드 베이스 내의 `<ServerIP>`, `<Port>`, `<Token>`, 그리고 `<VaultName>`의 위치에 이용자 각자의 적합도 정보를 적으십시오)*

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

## 🔗 클라이언트 및 클라이언트 플러그인 지원 목록

* Obsidian Fast Note Sync 플러그인
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool 미러 저장소](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* 서드 파티 연동 플러그인
  * [FastNodeSync-CLI ](https://github.com/Go1c/FastNodeSync-CLI) 파이썬(Python)의 기능과 백엔드 FNS WS 인터페이스 기술에 결합되어 제작된, 쌍방향 실시간 동기화를 달성한 CLI(명령 프롬프트 베이스)의 클라이언트 운영 체제 모듈입니다. 이는 디스플레이나 구체적 인터페이스 화면(GUI 환경)을 미처 지원하지 못해 구동 환경이 제한되는 서버 콘솔 환경(OpenClaw) 내에서 데스크톱 기반 Obsidian 동기화 능력을 확보하는데 최적화되었습니다.
