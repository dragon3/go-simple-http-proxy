.PHONY: dep dep-vendor-only clean test build build-image run

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

build-image:
	docker build --rm -t dragon3/simple-http-proxy .

push-image:
	docker push dragon3/simple-http-proxy

run:
	go run main.go
