#!/usr/bin/env bash

release_content=`cat files/torigoya-packages.list`
debug_content=`cat files/torigoya-packages.debug.list`

mode=""
if [ -e "/etc/apt/sources.list.d/torigoya-packages.list" ]; then
    content=`cat /etc/apt/sources.list.d/torigoya-packages.list`
    if [ "$content" == "$release_content" ]; then
        mode="release"
    elif [ "$content" == "$debug_content" ]; then
        mode="debug"
    else
        mode="unknown"
    fi
else
    echo "This environment was not set for torigoya_cage."
    exit 1
fi

if [ "$1" != "t" ]; then
    echo "== This environment is $mode mode."
else
    echo "$mode"
    exit 0
fi

case "$1" in
    "release")
        cp files/torigoya-packages.list /etc/apt/sources.list.d/torigoya-packages.list
        echo "Set to 'release' mode"
        ;;

    "debug")
        cp files/torigoya-packages.debug.list /etc/apt/sources.list.d/torigoya-packages.list
        echo "Set to 'debug' mode"
        ;;

    *)
        echo "An argument must be 'release' or 'debug'"
        ;;
esac
