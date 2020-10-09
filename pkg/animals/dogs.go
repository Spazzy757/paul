package animals

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

const (
	dogUrl = "https://api.thedogapi.com/v1/images/search"
)

type Dog struct {
	Url string `json:"url"`
}

func NewDogClient() *Client {
	return newClient(dogUrl)
}

// GetCat fetches a random cat url
func (cli *Client) GetDog() (Dog, error) {
	req, err := http.NewRequest("GET", cli.Url, nil)
	if err != nil {
		return Dog{}, errors.Wrap(err, "failed to build request")
	}

	resp, err := cli.HttpClient.Do(req)
	if err != nil {
		return Dog{}, errors.Wrap(err, "request failed")
	}

	var res []Dog
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return Dog{}, errors.Wrap(err, "unmarshaling failed")
	}

	return res[0], nil
}
