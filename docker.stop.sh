#!/usr/bin/env bash

CONTAINER_NAME=torigoya_cage

sudo docker ps --all | grep $CONTAINER_NAME
if [ $? == 0 ]; then
    # make container to stop
    sudo docker stop $CONTAINER_NAME

    # remove container
    sudo docker rm $CONTAINER_NAME || (echo "Failed to rm $CONTAINER_NAME"; exit -1)
fi
