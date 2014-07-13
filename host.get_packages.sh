echo "Torigoya: get go packages..."
GOPATH=`pwd` go get \
    gopkg.in/v1/yaml \
    github.com/jmcvetta/randutil \
    github.com/ugorji/go/codec \
    github.com/mattn/go-shellwords \
    || (echo "failed"; exit -1)
