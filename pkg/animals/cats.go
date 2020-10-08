package animals

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

const (
	catsUrl = "https://api.thecatapi.com/v1/images/search"
)

type Cat struct {
	Url string `json:"url"`
}

func NewCatClient() *Client {
	return newClient(catsUrl)
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
