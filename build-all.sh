#!/bin/sh

ROOT=`pwd`
NAME=tagbbs
RELEASE=`git describe --always --tag`

mkdir -p release
cd release

for os in windows linux darwin
do
    mkdir -p $NAME-$os-$RELEASE
    cd $NAME-$os-$RELEASE
    GOOS=$os go build github.com/thinxer/tagbbs/apibbsd
    GOOS=$os go build github.com/thinxer/tagbbs/sshbbsd
    cd ..
done

cp -r $ROOT/webui tagbbs-webui-$RELEASE
