#!/bin/sh
echo "docker images clean shell"
projectName=$(pwd | awk -F "/" '{print $NF}')
dockerrm=$(docker images | grep "${projectName}" | awk '{print $3}' | awk '!a[$0]++')

if [  -n "$dockerrm"  ]; then
    docker rmi -f ${dockerrm}
    echo "docker images ${projectName} clean OK"
fi

dockerrm=$(docker images | grep "none" | awk '{print $3}' | awk '!a[$0]++')
if [  -n "$dockerrm" ]; then
    docker rmi -f ${dockerrm}
    echo "docker images none clean OK"
fi
