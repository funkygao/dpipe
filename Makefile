build:
	mkdir -p var
	rm -f var/alser.lock
	go build

install:build
	go install 

clean:
	go clean

test:conf_test.go parser/all_test.go
	@go test -v
	@go test -v ./parser

run:build
	@rm -f var/alser.lock
	./alser -v -debug -test -tail -pprof var/cpu.prof

up:
	go get -u github.com/funkygao/alsparser
	go get -u github.com/funkygao/tail
	go get -u github.com/funkygao/alser

fmt:
	@gofmt -s -tabs=false -tabwidth=4 -w=true .

prof:build
	@go tool pprof alser var/cpu.prof

his:build
	./alser -c conf/alser.history.json

tail:build
	while true; do \
		./alser -c conf/alser.json -tail; \
	done
