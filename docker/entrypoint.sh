#!/bin/sh

# 检查环境变量
if [ -z "$P_NAME" ] || [ -z "$P_BIN" ]; then
    echo "Error: P_NAME or P_BIN not set"
    exit 1
fi

# 切换目录
cd "/${P_NAME}/" || { echo "Failed to cd to /${P_NAME}/"; exit 1; }

# 创建日志目录和文件
mkdir -p storage/logs || { echo "Failed to create logs dir"; exit 1; }
touch storage/logs/c.log || { echo "Failed to create c.log"; exit 1; }

# 备份旧日志，统一后缀为 .log
mv storage/logs/c.log "storage/logs/c_$(date '+%Y%m%d%H%M%S').log" || { echo "Failed to rename log"; exit 1; }

# 运行程序并记录日志到 c.log
"/${P_NAME}/${P_BIN}" run 2>&1 | tee storage/logs/c.log