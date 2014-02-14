#!/bin/sh

ROOT=`pwd`
NAME=tagbbs
RELEASE=`git describe --always --tag`

echo Release: $RELEASE

mkdir -p release
cd release

for os in windows linux darwin
do
    echo Building for $os...
    mkdir -p $NAME-$os-$RELEASE
    cd $NAME-$os-$RELEASE
    GOOS=$os go build github.com/thinxer/tagbbs/apibbsd
    GOOS=$os go build github.com/thinxer/tagbbs/sshbbsd
    cd ..
    ln -sfvn $NAME-$os-$RELEASE $NAME-$os
done

cp -r $ROOT/webui tagbbs-webui-$RELEASE
ln -sf tagbbs-webui-$RELEASE tagbbs-webui
