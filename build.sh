#! /bin/bash -e

#===========
# update
#===========
if [[ $1 = "-u" ]]; then
    go get -u github.com/funkygao/alser
fi

#===========
# build
#===========
cd $(dirname $0)
ID=$(git rev-parse HEAD | cut -c1-7)
go build -v -ldflags "-X main.BuildID $ID"

#===========
# show ver
#===========
./alser -version
