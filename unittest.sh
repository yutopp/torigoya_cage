#!/usr/bin/env bash

# echo "Torigoya test: build packages..."
# ./host.build.sh || exit -1

if [ "$1" != "" ]; then
    ext="-test.run $1"
else
    ext=""
fi

echo $ext

echo "Torigoya test: run unittest..."
sudo GOPATH=`pwd` go test -v -race yutopp/cage $ext
#sudo GOPATH=`pwd` go test -v -race -cover yutopp/cage $ext

#sudo GOPATH=`pwd` go test -v yutopp/cage -cpuprofile cpu.pprof -memprofile mem.pprof $ext \
#    && go tool pprof --text bin/cage cpu.pprof > cpu.txt \
#    && go tool pprof --text bin/cage mem.pprof > mem.txt
