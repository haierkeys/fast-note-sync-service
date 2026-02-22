# Docker 登录示例
# docker login --username=xxxxxx registry.cn-shanghai.aliyuncs.com

# 环境变量（可选）
# include .env
# export $(shell sed 's/=.*//' .env)

# -------------------------
# 项目 / 镜像 配置
# -------------------------
REPO = $(eval REPO := $$(shell go list -f '{{.ImportPath}}' .))$(value REPO)

DockerHubUser   = haierkeys
DockerHubName   = fast-note-sync-service

ReleaseTagPre   = release-v
DevelopTagPre   = develop-v

P_NAME          = fast-note-sync
P_BIN           = fast-note-sync-service

# -------------------------
# Git / 构建信息
# -------------------------
GitTag          = $(shell git describe --tags --abbrev=0)
GitVersion      = $(shell git log -1 --format=%h)
GitVersionDesc  = $(shell git log -1 --format=%s)
BuildTime       = $(shell date +%FT%T%z)

# LDFLAGS: 注入版本信息到二进制
LDFLAGS = -ldflags '-X ${REPO}/internal/app.Version=$(GitTag) -X "${REPO}/internal/app.GitTag=$(GitVersion)" -X ${REPO}/internal/app.BuildTime=$(BuildTime)'

# go 命令封装
gob = go build ${LDFLAGS}
gor = go run ${LDFLAGS}

# 编译相关
CGO = CGO_ENABLED=0

rootDir = $(shell pwd)
buildDir = $(rootDir)/build

# -------------------------
# PHONY 目标
# -------------------------
.PHONY: all  build-all run test clean \
        push-online push-dev \
        build-macos-amd64 build-macos-arm64 build-linux-amd64 \
        build-linux-arm64 build-windows-amd64 gox-linux gox-all \
		docs fmt update air dev ver gen sup

# 默认目标
all: test build-all

# -------------------------
# 简单目标
# -------------------------
sup:
	node scripts/process_support_csv.js
	python3 scripts/process_support.py
	node scripts/gen_support_md.js

sup-md:
	node scripts/gen_support_md.js
test:
	@echo ${REPO}

dev:
	air -c ./scripts/.air.toml

air:
	air -c ./scripts/.air.toml

fmt:
	go fmt ./...

update:
	go get -u ./...

# 更新版本脚本调用
ver:
	@node ./scripts/update-version.js $(filter-out $@,$(MAKECMDGOALS))

# 捕获 ver 后面的参数，防止 make 将其视为目标
%:
	@:

gen:
	go run -v ./cmd/gorm_gen/gen.go -type sqlite -dsn storage/database/db.sqlite3
	go run -v ./cmd/model_gen/gen.go

docs:
	go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o ./docs --parseDependency --parseInternal

# 运行
run:
#	$(call checkStatic)
	$(call init)
	$(gor) -v $(rootDir)

clean:
	rm -rf $(buildDir)

# -------------------------
# 构建集合
# -------------------------
build-all:
#	$(call checkStatic)
	$(MAKE) build-macos-amd64
	$(MAKE) build-macos-arm64
	$(MAKE) build-linux-amd64
	$(MAKE) build-linux-arm64
	$(MAKE) build-windows-amd64

# macOS
build-macos-amd64:
	$(CGO) GOOS=darwin GOARCH=amd64 $(gob) -o $(buildDir)/darwin_amd64/${P_BIN} $(bin) -v $(rootDir)

build-macos-arm64:
	$(CGO) GOOS=darwin GOARCH=arm64 $(gob) -o $(buildDir)/darwin_arm64/${P_BIN} -v $(rootDir)

# Linux
build-linux-amd64:
# CGO_ENABLED=1 CC=musl-gcc  GOOS=linux GOARCH=amd64 $(gob) -o $(buildDir)/linux_amd64/${P_BIN} -v $(rootDir)
	$(CGO) GOOS=linux GOARCH=amd64 $(gob) -o $(buildDir)/linux_amd64/${P_BIN} -v $(rootDir)

build-linux-arm64:
	$(CGO) GOOS=linux GOARCH=arm64 $(gob) -o $(buildDir)/linux_arm64/${P_BIN} -v $(rootDir)

# Windows
build-windows-amd64:
# CGO_ENABLED=0 CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="x86_64-w64-mingw32-gcc -fno-stack-protector -D_FORTIFY_SOURCE=0 -lssp" $(gob) -o $(bin).exe -v $(rootDir)
	$(CGO) GOOS=windows GOARCH=amd64 $(gob) -o $(buildDir)/windows_amd64/${P_BIN}.exe -v $(rootDir)

# gox 辅助
gox-linux:
	$(CGO) gox ${LDFLAGS} -osarch="linux/amd64 linux/arm64" -output="$(buildDir)/{{.OS}}_{{.Arch}}/${P_BIN}"

gox-all:
	$(CGO) gox ${LDFLAGS} -osarch="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64" -output="$(buildDir)/{{.OS}}_{{.Arch}}/${P_BIN}"


# -------------------------
# Docker 发布
# -------------------------
push-online: build-linux
	$(call dockerImageClean)
	docker build --platform linux/amd64 -t $(DockerHubUser)/$(DockerHubName):latest -f docker/Dockerfile .
	docker tag  $(DockerHubUser)/$(DockerHubName):latest $(DockerHubUser)/$(DockerHubName):$(ReleaseTagPre)$(GitTag)
	docker push $(DockerHubUser)/$(DockerHubName):$(ReleaseTagPre)$(GitTag)
	docker push $(DockerHubUser)/$(DockerHubName):latest

push-dev: build-linux
	$(call dockerImageClean)
	docker build --platform linux/amd64 -t $(DockerHubUser)/$(DockerHubName):dev-latest -f docker/Dockerfile .
	docker tag $(DockerHubUser)/$(DockerHubName):dev-latest $(DockerHubUser)/$(DockerHubName):$(DevelopTagPre)$(GitTag)
	docker push $(DockerHubUser)/$(DockerHubName):$(DevelopTagPre)$(GitTag)
	docker push $(DockerHubUser)/$(DockerHubName):dev-latest

# -------------------------
# 代码片段（定义）
# -------------------------
define dockerImageClean
	@echo "docker Image Clean"
	bash docker_image_clean.sh
endef

define init
	@echo "Build Init"
endef
