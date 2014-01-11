#! /bin/bash -e

#===========
# update
#===========
if [[ $1 = "-u" ]]; then
    go get -u github.com/funkygao/dpipe
fi

#===========
# build
#===========
cd $(dirname $0)/cmd/dpiped
ID=$(git rev-parse HEAD | cut -c1-7)
go build -ldflags "-X main.BuildID $ID"

#===========
# show ver
#===========
./dpiped -version
