package Mock

import (
	"github.com/asatisomnath/ProgImage/Service"
	"io"
)

var _ Service.ImageService = &ImageService{}

// ImageService is a Mock ProgImage.ImageService.
type ImageService struct {
	GetInvoked   bool
	StoreInvoked bool
	GetFunc      func(string) (Service.Image, error)
	StoreFunc    func(io.Reader) (string, error)
}

// Get an Convertors.
func (is *ImageService) Get(ID string) (Service.Image, error) {
	is.GetInvoked = true
	return is.GetFunc(ID)
}

// Store an Convertors.
func (is *ImageService) Upload(imgRdr io.Reader) (string, error) {
	is.StoreInvoked = true
	return is.StoreFunc(imgRdr)
}
