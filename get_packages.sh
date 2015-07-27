echo "Torigoya: get go packages..."
GOPATH=`pwd` go get \
    gopkg.in/v1/yaml \
    github.com/ugorji/go/codec \
    github.com/jmcvetta/randutil \
    || (echo "failed"; exit -1)
