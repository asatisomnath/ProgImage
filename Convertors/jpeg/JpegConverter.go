package jpeg

import (
	"image"
	_ "image/gif" // register Convertors type, do not remove
	"image/jpeg"
	_ "image/png" // register Convertors type, do not remove
	"io"

	primage "github.com/asatisomnath/ProgImage/Convertors"

)

// Transformer implements ProgImage.ImageTypeTransformer to convert a ProgImage.Image to jpeg format.
var Converter = primage.Converter{
	Name:        "jpeg",
	ContentType: "Convertors/jpeg",
	Encoder:     DefaultJpegEncode,
}

// DefaultJpegEncode performs jpeg encoding with default values for options.
func DefaultJpegEncode(w io.Writer, m image.Image) error {
	return jpeg.Encode(w, m, nil)
}
