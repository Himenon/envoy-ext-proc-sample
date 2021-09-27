package share

import (
	"io/ioutil"
	"net"
	"os"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

func CreateTcpListener(address string) net.Listener {
	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		logrus.Fatal("failed to listen: %v", err)
	}
	return tcpListener
}

func CreateGrpcServer(concurrent uint32) *grpc.Server {
	opts := []grpc.ServerOption{grpc.MaxConcurrentStreams(concurrent)}
	opts = append(opts)
	return grpc.NewServer(opts...)
}

func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func GetHeaderValue(headers []*core.HeaderValue, key string) (string, bool) {
	result := ""
	ok := false
	for _, headerValue := range headers {
		if headerValue.Key == key {
			result = headerValue.Value
			ok = true
			break
		}
	}
	return result, ok
}

func ReadFile(htmlFilePath string) string {
	bytes, err := ioutil.ReadFile(htmlFilePath)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
