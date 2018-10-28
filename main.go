package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type config struct {
	Debug bool
	Addr  string `default:":9000"`
}

type proxyHandler struct {
	logger *zap.Logger
}

func main() {
	var c config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatalf("failed to process env vars: %+v", err)
	}

	logger, err := newLogger(c)
	if err != nil {
		log.Fatalf("failed to create logger: %+v", err)
	}

	proxyHandler := newProxyHandler(logger)

	server := &http.Server{
		Addr: c.Addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				proxyHandler.handleConnect(w, r)
			} else {
				proxyHandler.handleHTTP(w, r)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	logger.Info(fmt.Sprintf("Starting proxy server: addr=%s", c.Addr))

	log.Fatal(server.ListenAndServe())
}

func newProxyHandler(logger *zap.Logger) *proxyHandler {
	return &proxyHandler{
		logger: logger,
	}
}

func (h *proxyHandler) handleConnect(w http.ResponseWriter, r *http.Request) {
	h.logger.Info(fmt.Sprintf("handling HTTP CONNECT: %v", r))
	targetConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	go copyIO(targetConn, clientConn)
	go copyIO(clientConn, targetConn)
}

func (h *proxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info(fmt.Sprintf("handling HTTP request: %v", r))
	res, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer res.Body.Close()
	copyHeader(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
}

func copyIO(dst io.WriteCloser, src io.ReadCloser) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func newLogger(c config) (*zap.Logger, error) {
	var level zapcore.Level
	if c.Debug {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}
	zc := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	return zc.Build()
}
