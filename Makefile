build:
	go build

install:
	go install 

clean:
	go clean

test:
	@go test -v

run:build
	@rm -f var/alser.lock
	./alser -v -debug
