#!/bin/bash
rootdir="$( cd "$( dirname "$0"  )" && pwd  )"
webroot=${rootdir}

ps aux|grep goapi_starfission|awk '{print $2}'|xargs kill -9
