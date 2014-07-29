#!/usr/bin/env bash

cwd=`pwd`

./docker.stop.sh &&
./docker.build.sh &&
echo "start container => " &&
sudo docker run \
    -v $cwd/files/proc_profiles_for_core_test/:/opt/cage/files/proc_profiles_for_core_test/ \
    -v $cwd/host.unittest.sh:/opt/cage/host.unittest.sh \
    --name torigoya_cage \
    --workdir /opt/cage \
    --privileged \
    torigoya/cage \
    ./host.unittest.sh
