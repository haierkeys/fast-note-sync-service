#!/bin/bash

ProjectRegistry="registry.cn-shanghai.aliyuncs.com/xxx/xxxxx"
ProjectPath=`pwd`
ProjectName="xxxxx"

Usage() {
    echo "Usage:"
    echo "test.sh [-t Git tag]"
    echo "Description:"
    exit
}

while getopts ':t:h:' OPT; do
    case $OPT in
        t) TAG="$OPTARG";;
        h) Usage;;
        ?) Usage;;
    esac
done

if [ ${TAG} ];then
    docker pull $ProjectRegistry:$TAG

    echo "Stop "$ProjectName

    docker stop $ProjectName
    #docker rm -v
    docker rm -f $ProjectName
    echo "Start new xxxxx"
    docker run -tid --name $ProjectName \
        -p 8000:8000 -p 8001:8001 -p 8002:8002 \
        -v $ProjectPath/storage/:/api/storage/ \
        -v $ProjectPath/configs/:/api/configs/ \
        $ProjectRegistry:$TAG
fi


