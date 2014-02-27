#! /bin/bash -e

if [[ $1 = "-loc" ]]; then
    find . -name '*.go' | xargs wc -l | sort -n
    exit
fi

cd $(dirname $0)/cmd/dpiped
ID=$(git rev-parse HEAD | cut -c1-7)
if [[ $1 = "-linux" ]]; then
    #=======================================
    # to enable cross compiling, you need to 
    #=======================================
    # cd $GOROOT/src; CGO_ENABLED=0 GOOS=linux GOARCH=386 ./make.bash
    CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-X main.BuildID $ID"
else
    #go build -race -v -ldflags "-X main.BuildID $ID"
    go build -ldflags "-X main.BuildID $ID"
fi

#=========
# show ver
#=========
./dpiped -version
