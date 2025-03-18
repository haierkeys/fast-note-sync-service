#!/bin/bash
rootdir="$( cd "$( dirname "$0"  )" && pwd  )"
webroot=${rootdir}

cd ${webroot}
cp -rf ./new/apiRun ./apiRun
chmod +x ./apiRun
ps aux|grep goapi_starfission|awk '{print $2}'|xargs kill -9
${webroot}/apiRun
