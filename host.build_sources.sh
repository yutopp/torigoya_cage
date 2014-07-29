echo "Torigoya: build sources..."
GOPATH=`pwd` go build -o bin/cage.server yutopp/cage.server &&
GOPATH=`pwd` go build -o bin/cage.callback yutopp/cage.callback &&
make -f Makefile.posix
