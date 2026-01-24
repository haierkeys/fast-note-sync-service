#!/bin/sh
echo "docker images clean shell"

projectName=$(basename "$(pwd)")

# 删除匹配 repository 名称（任意 tag）的镜像
dockerrm=$(docker images --filter "reference=${projectName}:*" -q | sort -u)
if [ -n "$dockerrm" ]; then
    echo "$dockerrm" | xargs -r docker rmi -f
    echo "docker images ${projectName} clean OK"
fi

# 删除 dangling (none) 镜像
dockerrm=$(docker images -f "dangling=true" -q | sort -u)
if [ -n "$dockerrm" ]; then
    echo "$dockerrm" | xargs -r docker rmi -f
    echo "docker images none clean OK"
fi
