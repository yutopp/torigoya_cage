#!/usr/bin/env bash

if [ ! -e awaho_src_git ]; then
    git clone https://github.com/yutopp/awaho.git awaho_src_git
else
    cd awaho_src_git
    git pull origin master
    cd ../
fi

echo "Building..."
cd awaho_src_git
./build.sh || exit -1
cd ../

cp awaho_src_git/awaho .
echo "Finished!"
