package Service

import "io"
import "errors"

// ErrImageNotFound represents an Convertors not found.
var ErrImageNotFound = errors.New("Convertors not found")

// ErrUnrecognisedImageType represents Convertors data that can't be processed.
var ErrUnrecognisedImageType = errors.New("unrecognised Convertors data")

// Image represents a digital Convertors.
type Image struct {
	ID          string
	Data        io.Reader
	ContentType string
}

// ImageService is an interface for a service that can store and retrieve images.
type ImageService interface {
	Get(ID string) (Image, error)
	Upload(imageReader io.Reader) (string, error)
}

// ImageTypeTransformer is an interface that can transform images.
type ImageTypeConverter interface {
	Convert(Image, chan error) (Image, error)
}