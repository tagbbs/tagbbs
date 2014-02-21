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
    GOOS=$os go build -ldflags "-X github.com/tagbbs/tagbbs.version $RELEASE" github.com/tagbbs/tagbbs/apibbsd
    GOOS=$os go build -ldflags "-X github.com/tagbbs/tagbbs.version $RELEASE" github.com/tagbbs/tagbbs/sshbbsd
    GOOS=$os go build -ldflags "-X github.com/tagbbs/tagbbs.version $RELEASE" github.com/tagbbs/tagbbs/fsck
    GOOS=$os go build -ldflags "-X github.com/tagbbs/tagbbs.version $RELEASE" github.com/tagbbs/tagbbs/bench
    cd ..
    ln -sfvn $NAME-$os-$RELEASE $NAME-$os
    tar zcf $NAME-$os-$RELEASE.tgz $NAME-$os-$RELEASE
done
