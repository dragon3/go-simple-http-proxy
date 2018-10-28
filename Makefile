.PHONY: dep dep-vendor-only clean test build run

dep:
	dep ensure -v

dep-vendor-only:
	dep ensure -v -vendor-only

clean:
	rm -rf bin/

test:
	go test -v ./...

build:
	CGO_ENABLED=0 go build -o bin/simple-http-proxy

run:
	go run main.go
