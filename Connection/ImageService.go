package Connection

import (
	"encoding/json"
	"github.com/asatisomnath/ProgImage/Service"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type GetterDoer interface {
	Get(string) (*http.Response, error)
	Do(r *http.Request) (*http.Response, error)
}

// ImageService is a ProgImage.ImageService that makes requests to a Connection server.
type ImageService struct {
	BaseURL string
	Client  GetterDoer
}

var _ Service.ImageService = ImageService{}

// Get the Convertors for the given ID.
func (is ImageService) Get(ID string) (Service.Image, error) {
	resp, err := is.Client.Get(is.BaseURL + "/image/" + ID)
	ret := Service.Image{}
	if err != nil {
		return ret, errors.Wrap(err, "unable to make get request")
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return ret, Service.ErrImageNotFound
		}
		return ret, errors.Errorf("unknown error getting image, status code %d", resp.StatusCode)
	}

	ret.ID = ID
	ret.ContentType = resp.Header.Get("ContentType")
	ret.Data = resp.Body
	return ret, nil
}

// Store an Convertors.
func (is ImageService) Upload(imgRdr io.Reader) (string, error) {
	req, err := http.NewRequest("POST", is.BaseURL+"/image/create", imgRdr)
	if err != nil {
		return "", errors.Wrap(err, "unable to create new Connection request")
	}
	resp, err := is.Client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "unable to make post request")
	}
	if resp.StatusCode != http.StatusCreated {
		return "", errors.Errorf("unknown error creating new image, status code %d", resp.StatusCode)
	}

	rd := new(respData)
	if err := json.NewDecoder(resp.Body).Decode(rd); err != nil {
		return "", errors.Wrap(err, "error decoding resp")
	}

	return rd.ID, nil
}

type respData struct {
	ID string
}
