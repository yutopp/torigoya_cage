#!/usr/bin/env bash

echo "Torigoya test: build packages..."
./host.build.sh || exit -1

echo "Torigoya test: run unittest..."
sudo GOPATH=`pwd` go test yutopp/torigoya/cage
