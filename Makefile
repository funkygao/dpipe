test:
	@go test -v

run:
	@rm -f var/alser.lock
	./alser -v -debug -test -tail -pprof var/cpu.prof

fmt:
	@gofmt -s -tabs=false -tabwidth=4 -w=true .

prof:
	@go tool pprof alser var/cpu.prof

his:
	./alser -c conf/alser.history.json

tail:
	while true; do \
		./alser -c conf/alser.json -tail; \
	done
