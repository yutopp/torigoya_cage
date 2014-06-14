GOPATH=`pwd` go get gopkg.in/v1/yaml
GOPATH=`pwd` go get github.com/jmcvetta/randutil
GOPATH=`pwd` go get github.com/ugorji/go/codec

GOPATH=`pwd` go build -o bin/cage yutopp/cage
GOPATH=`pwd` go build -o bin/cage.callback yutopp/cage.callback
make -f Makefile.posix

# sudo docker build -t torigoya/cage
