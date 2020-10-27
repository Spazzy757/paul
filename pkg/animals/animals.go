package animals

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

const (
	catsUrl = "https://api.thecatapi.com/v1/images/search"
	dogUrl  = "https://api.thedogapi.com/v1/images/search"
)

//Animal defines the struct value for the animal
type Animal struct {
	Url string `json:"url"`
}

//NewCatClient issues a new Cat client
func NewCatClient() *Client {
	return newClient(catsUrl)
}

//NewDogClient returns a client of type Dog
func NewDogClient() *Client {
	return newClient(dogUrl)
}

// GetLink fetches a random cat or dog url
func (cli *Client) GetLink() (Animal, error) {
	req, err := http.NewRequest("GET", cli.Url, nil)
	if err != nil {
		return Animal{}, errors.Wrap(err, "failed to build request")
	}

	resp, err := cli.HttpClient.Do(req)
	if err != nil {
		return Animal{}, errors.Wrap(err, "request failed")
	}

	var res []Animal
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return Animal{}, errors.Wrap(err, "unmarshaling failed")
	}

	return res[0], nil
}
