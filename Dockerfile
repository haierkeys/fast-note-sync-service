FROM woahbase/alpine-glibc:latest
MAINTAINER HaierKeys <haierkeys@gmail.com>
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG BUILD_DATE
ARG GIT_COMMIT

ARG VERSION=${VERSION}
ARG BUILD_DATE=${BUILD_DATE}
ARG GIT_COMMIT=${GIT_COMMIT}



LABEL name="obsidian-better-sync-service"
LABEL version=${VERSION}
LABEL description="Provide image resizing, cropping, upload/download, and cloud storage features for Obsidian CIAU."
LABEL maintainer="HaierKeys <haierkeys@gmail.com>"


LABEL org.opencontainers.image.title="Obsidian Better Sync Service"
LABEL org.opencontainers.image.created=${BUILD_DATE}
LABEL org.opencontainers.image.authors="HaierKeys <haierkeys@gmail.com>"
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.description="Provide image resizing, cropping, upload/download, and cloud storage features for Obsidian CIAU."
LABEL org.opencontainers.image.url="https://github.com/haierkeys/obsidian-better-sync-service"
LABEL org.opencontainers.image.source="https://github.com/haierkeys/obsidian-better-sync-service"
LABEL org.opencontainers.image.documentation="https://raw.githubusercontent.com/haierkeys/obsidian-better-sync-service/refs/heads/main/README.md"
LABEL org.opencontainers.image.revision=${GIT_COMMIT}
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="HaierKeys"



ENV TZ=Asia/Shanghai
ENV P_NAME=better-sync
ENV P_BIN=better-sync-service
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --update add libstdc++ curl ca-certificates bash curl gcompat tzdata && \
    cp /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo ${TZ} > /etc/timezone && \
    rm -rf  /tmp/* /var/cache/apk/*

EXPOSE 9000 9001
RUN mkdir -p /${P_NAME}/
VOLUME /${P_NAME}/config
VOLUME /${P_NAME}/storage
COPY ./build/${TARGETOS}_${TARGETARCH}/${P_BIN} /${P_NAME}/

# 将脚本复制到容器中
COPY entrypoint.sh /entrypoint.sh

# 给脚本执行权限
RUN chmod +x /entrypoint.sh

# 使用 ENTRYPOINT 执行脚本
ENTRYPOINT ["/entrypoint.sh"]