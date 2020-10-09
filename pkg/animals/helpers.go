package animals

import (
	"net/http"
	"time"
)

//Client defines the client struct
type Client struct {
	HttpClient *http.Client
	Url        string
}

func newClient(url string) *Client {
	client := Client{
		HttpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		// Place holder url
		Url: url,
	}

	return &client
}
