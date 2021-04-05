package gif

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/pkg/errors"
)

const (
	giphyUrl = "https://api.giphy.com/v1/gifs/search"
)

//Client defines the client struct
// TODO Move this into the helpers
type Client struct {
	HttpClient *http.Client
	Url        string
}

func newClient(url string) *Client {
	apiKey := helpers.GetEnv("GIPHY_API_KEY", "")
	client := Client{
		HttpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		// Place holder url
		Url: url + "?api_key=" + apiKey,
	}

	return &client
}

// gipyImageDataExtended is the extended image data struct
// with additional video fields
type gipyImageDataExtended struct {
	Url, Width, Height, Size, Mp4, Mp4_size, Webp, Webp_size string
}

// giphyDataArray is a struct holding multiple API result entries
type giphyDataArray struct {
	Data []struct {
		Images struct {
			Fixed_height gipyImageDataExtended
		}
	}
}

//NewCatClient issues a new Cat client
func NewGifClient() *Client {
	return newClient(giphyUrl)
}

// GetLink fetches a random cat or dog url
func (cli *Client) GetLink(search string) (string, error) {
	url := cli.Url + "&q=" + search
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to build request")
	}

	resp, err := cli.HttpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "request failed")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "request failed")
	}

	var data giphyDataArray
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", errors.Wrap(err, "unmarshaling failed")
	}
	i := rand.Intn(len(data.Data))
	gifUrl := data.Data[i].Images.Fixed_height.Url

	return gifUrl, nil
}
