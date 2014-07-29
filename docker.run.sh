#!/usr/bin/env bash

cwd=`pwd`

echo "Torigoya system_test: use torigoya-packages.system_test.list..."
packages_list_path=$cwd/files/torigoya-packages.list

echo "Torigoya run: use  $packages_list_path"

# ========================================
./docker.stop.sh &&
./docker.build.sh &&
echo "start container => " &&
sudo docker run \
     --expose 23432 \
     -p 23432:23432 \
     -v $cwd/config.yml:/opt/cage/config.yml \
     -v $cwd/torigoya_proc_profiles:/opt/cage/proc_profiles \
     -v $packages_list_path:/etc/apt/sources.list.d/torigoya-packages.list \
     --name torigoya_cage \
     --workdir /opt/cage \
     --privileged \
     torigoya/cage \
     bin/cage.server --mode release
