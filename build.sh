#!/usr/bin/env bash

cd _awaho
./build.sh || exit -1
cd ../

./build_cage.sh
