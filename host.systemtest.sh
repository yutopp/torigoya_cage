#!/usr/bin/env bash

if [ "$1" != "skip" ]; then
    echo "Torigoya test: build packages..."
    ./host.build.sh || exit -1
fi

echo "Torigoya test: run system test..."
sudo bundle exec ruby test/system_test.rb
