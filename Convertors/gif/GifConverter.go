package gif

import (
	"image"
	"image/gif"
	_ "image/jpeg" // import to register Convertors type
	_ "image/png"  // import to register Convertors type
	"io"
	primage "github.com/asatisomnath/ProgImage/Convertors"

)

// Transformer implements ProgImage.ImageTypeTransformer to convert a ProgImage.Image to png format.
var Converter = primage.Converter{
	Name:        "gif",
	ContentType: "Convertors/gif",
	Encoder:     DefaultGifEncode,
}

// DefaultGifEncode performs fig encoding with default values for options.
func DefaultGifEncode(w io.Writer, m image.Image) error {
	return gif.Encode(w, m, nil)
}
