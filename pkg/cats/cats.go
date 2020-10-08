package cats

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	catsUrl = "https://api.thecatapi.com/v1/images/search"
)

type Client struct {
	HttpClient *http.Client
	Url        string
}

func NewClient() *Client {
	client := Client{
		HttpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		Url: catsUrl,
	}

	return &client
}

type Cat struct {
	Url string `json:"url"`
}

// GetCat fetches a random cat url
func (cli *Client) GetCat() (Cat, error) {

	req, err := http.NewRequest("GET", cli.Url, nil)
	if err != nil {
		return Cat{}, errors.Wrap(err, "failed to build request")
	}

	resp, err := cli.HttpClient.Do(req)
	if err != nil {
		return Cat{}, errors.Wrap(err, "request failed")
	}

	var res []Cat
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return Cat{}, errors.Wrap(err, "unmarshaling failed")
	}

	return res[0], nil
}
