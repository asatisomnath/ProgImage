package imageConvertors

import (
	"fmt"
	"github.com/asatisomnath/ProgImage/Service"
	"image"
	"io"
	"log"

	"github.com/pkg/errors"
)

var _ Service.ImageTypeConverter = Converter{}

// Transformer enables ProgImage.ImageTypeTransformer implementations to be created easily avoiding code duplication.
type Converter struct {
	ContentType string
	Encoder     func(io.Writer, image.Image) error
	Name        string
}

// Transform the given Convertors to the desired format.
func (t Converter) Convert(img Service.Image, ec chan error) (Service.Image, error) {
	if img.ContentType == t.ContentType {
		ec <- nil
		return img, nil
	}
	ret := Service.Image{}
	i, _, err := image.Decode(img.Data)
	if err != nil {
		return ret, errors.Wrap(err, fmt.Sprintf("unable to decode %s Convertors", t.Name))
	}

	r, w := io.Pipe()
	go func() {
		if err := t.Encoder(w, i); err != nil {
			ec <- errors.Wrap(err, fmt.Sprintf("unable to encode %s Convertors", t.Name))
			closeErr := w.Close()
			if closeErr != nil {
				log.Println("error closing pipe (unable to encode)", closeErr.Error())
			}
			return
		}
		ec <- nil
		closeErr := w.Close()
		if closeErr != nil {
			log.Println("error closing pipe", closeErr.Error())
		}
	}()

	ret.ID = img.ID
	ret.ContentType = t.ContentType
	ret.Data = r
	return ret, nil
}
