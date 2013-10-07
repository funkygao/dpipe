build:prepare
	go build

install:
	go install 

clean:
	go clean

test:conf_test.go parser/all_test.go
	@go test -v
	@go test -v ./parser

run:build
	@rm -f var/alser.lock
	./alser -v -debug -test

prepare:
	@go get -u github.com/funkygao/alser/parser

doc:
	@go doc github.com/funkygao/alser/parser

fmt:
	@gofmt -s -tabs=false -tabwidth=4 -w=true .
