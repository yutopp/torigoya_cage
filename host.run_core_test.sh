#!/usr/bin/env bash

echo "Torigoya test: build packages..."
./host.build.sh || exit -1

echo "Torigoya test: run core test..."
sudo GOPATH=`pwd` go test -v yutopp/torigoya/cage
