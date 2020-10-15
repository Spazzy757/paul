package helpers

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
)

/*
GetEnv looks up an env key or returns a default
*/
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

//MockHTTPClient A helper to Mock out Http Servers for testing
func MockHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewTLSServer(handler)
	cli := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		},
	}

	return cli, s.Close
}
