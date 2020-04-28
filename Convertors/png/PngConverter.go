package png

import (
	_ "image/gif"  // register Convertors type, do not remove
	_ "image/jpeg" // register Convertors type, do not remove
	"image/png"

	primage "github.com/asatisomnath/ProgImage/Convertors"
)

// Transformer implements ProgImage.ImageTypeTransformer to convert a ProgImage.Image to png format.
var Converter = primage.Converter{
	Name:        "png",
	ContentType: "Convertors/png",
	Encoder:     png.Encode,
}
