#!/usr/bin/env bash

cwd=`pwd`

if [ "$1" == "remote" ]; then
    echo "Torigoya system_test: use torigoya-packages.system_test.list..."
    packages_list_path=$cwd/files/torigoya-packages.list
    extra_commands=""

else
    echo "Torigoya system_test: use local apt repository..."
    host_apt_path="$cwd/torigoya_factory/apt_repository/"

    if [ ! -e $host_apt_path ]; then
        echo "Directory '$host_apt_path' was not found."
        echo "  Pleace clone 'torigoya_factory' repository into this directory."
        echo "  e.g. execute 'git clone https://github.com/yutopp/torigoya_factory.git'"
        exit -1
    fi

    #
    tmp_file=`mktemp`
    deb_path="deb file:///opt/cage/apt_repository/ trusty main"
    echo "writing...: $deb_path"
    echo $deb_path > $tmp_file

    #
    packages_list_path=$tmp_file
    extra_commands="-v $host_apt_path:/opt/cage/apt_repository/"
fi

echo "Torigoya system_test: use  $packages_list_path"
echo "Torigoya system_test: exec $extra_commands"

# ========================================
./docker.stop.sh &&
./docker.build.sh &&
echo "start container => " &&
sudo docker run \
     --expose 49800 \
     -p 49800:23432 \
     -v $cwd/config.yml:/opt/cage/config.yml \
     -v $cwd/torigoya_proc_profiles:/opt/cage/proc_profiles \
     -v $cwd/files/packages:/usr/local/torigoya \
     -v $packages_list_path:/etc/apt/sources.list.d/torigoya-packages.list \
     $extra_commands \
     --name torigoya_cage \
     --workdir /opt/cage \
     --privileged \
     torigoya/cage \
     bin/cage.server --mode system_test_mode
