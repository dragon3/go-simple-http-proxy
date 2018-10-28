FROM golang:1.11.1-alpine AS builder
RUN apk add --no-cache make git
RUN go get github.com/golang/dep/cmd/dep
COPY . /go/src/github.com/dragon3/go-simple-http-proxy
WORKDIR /go/src/github.com/dragon3/go-simple-http-proxy
RUN make dep-vendor-only build

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/dragon3/go-simple-http-proxy/bin/simple-http-proxy /simple-http-proxy
CMD ["/simple-http-proxy"]
