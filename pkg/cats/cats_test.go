package cats

import (
	"context"
	"crypto/tls"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
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

func TestClientGetCat(t *testing.T) {
	catAPIResponse := `[
        {
            "breeds":[],
            "id":"40g",
            "url":"https://cdn2.thecatapi.com/images/40g.jpg",
            "width":640,
            "height":426
         }
    ]`
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(catAPIResponse))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	client := NewClient()
	client.httpClient = httpClient
	client.Url = "https://example.com"
	t.Run("Test Get Cat Returns Cat Type", func(t *testing.T) {
		cat, err := client.GetCat()
		assert.Nil(t, err)
		assert.Equal(t, cat.Url, "https://cdn2.thecatapi.com/images/40g.jpg")
	})
}
