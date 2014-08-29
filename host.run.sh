#!/usr/bin/env bash

if [ $1 == "skip" ]; then
    sudo bin/cage.server
else
    sudo bin/cage.server --update true
fi
