#!/usr/bin/env bash

echo "Torigoya test: build packages..."
./host.build.sh || exit -1

echo "Torigoya test: run unittest..."
sudo GOPATH=`pwd` go test -v yutopp/cage -cpuprofile cpu.pprof -memprofile mem.pprof \
    && go tool pprof --text bin/cage cpu.pprof > cpu.txt \
    && go tool pprof --text bin/cage mem.pprof > mem.txt
