test:
	@go test -v
	cd parser; go test -v

build:
	go build

run:build
	@rm -f var/alser.lock
	./alser -v -debug -test -tail -pprof var/cpu.prof

fmt:
	@gofmt -s -tabs=false -tabwidth=4 -w=true .

prof:build
	@go tool pprof alser var/cpu.prof

his:build
	@rm -f var/*
	./alser -c conf/alser.history.json

tail:build
	while true; do \
		rm -f var/*; \
		./alser -c conf/alser.json -tail; \
		sleep 3; \
	done
