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
    GOOS=$os go build -ldflags "-X github.com/thinxer/tagbbs.version $RELEASE" github.com/thinxer/tagbbs/apibbsd
    GOOS=$os go build -ldflags "-X github.com/thinxer/tagbbs.version $RELEASE" github.com/thinxer/tagbbs/sshbbsd
    cd ..
    ln -sfvn $NAME-$os-$RELEASE $NAME-$os
    tar zcf $NAME-$os-$RELEASE.tgz $NAME-$os-$RELEASE
done

cp -r $ROOT/webui tagbbs-webui-$RELEASE
ln -sfvn tagbbs-webui-$RELEASE tagbbs-webui
tar zcf tagbbs-webui-$RELEASE.tgz tagbbs-webui-$RELEASE
