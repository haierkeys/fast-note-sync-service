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



LABEL name="fast-note-sync-service"
LABEL version=${VERSION}
LABEL description="Provide image resizing, cropping, upload/download, and cloud storage features for Obsidian CIAU."
LABEL maintainer="HaierKeys <haierkeys@gmail.com>"


LABEL org.opencontainers.image.title="Fast Note Sync Service"
LABEL org.opencontainers.image.created=${BUILD_DATE}
LABEL org.opencontainers.image.authors="HaierKeys <haierkeys@gmail.com>"
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.description="Provide image resizing, cropping, upload/download, and cloud storage features for Obsidian CIAU."
LABEL org.opencontainers.image.url="https://github.com/haierkeys/fast-note-sync-service"
LABEL org.opencontainers.image.source="https://github.com/haierkeys/fast-note-sync-service"
LABEL org.opencontainers.image.documentation="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/main/README.md"
LABEL org.opencontainers.image.revision=${GIT_COMMIT}
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="HaierKeys"



ENV TZ=Asia/Shanghai
ENV P_NAME=fast-note-sync
ENV P_BIN=fast-note-sync-service
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add --no-progress libstdc++ curl ca-certificates bash gcompat tzdata && \
    cp /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo ${TZ} > /etc/timezone && \
    mkdir -p /${P_NAME}/

EXPOSE 9000 9001
VOLUME /${P_NAME}/config
VOLUME /${P_NAME}/storage
COPY ./build/${TARGETOS}_${TARGETARCH}/${P_BIN} /${P_NAME}/
COPY ./docker/entrypoint.sh  /entrypoint.sh

RUN chmod +x /entrypoint.sh /${P_NAME}/${P_BIN}

ENTRYPOINT ["/entrypoint.sh"]