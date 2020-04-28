package jpeg_test

import (
	"github.com/asatisomnath/ProgImage/Service"
	"image"
	"os"
	"testing"

	"github.com/asatisomnath/ProgImage/Convertors/jpeg"
)

var fileTests = []struct {
	Name        string
	Path        string
	ContentType string
}{
	{Name: "png", Path: "../../testimages/test.png", ContentType: "Convertors/png"},
	{Name: "gif", Path: "../../testimages/test.gif", ContentType: "Convertors/gif"},
	{Name: "jpg", Path: "../../testimages/test.jpg", ContentType: "Convertors/jpeg"},
}

func TestTransformPNG(t *testing.T) {
	for _, item := range fileTests {
		t.Run(item.Name, func(t *testing.T) {
			fp, err := os.Open(item.Path)
			if err != nil {
				t.Fatal(err)
			}
			defer fp.Close()

			img := Service.Image{
				ID:          item.Name,
				ContentType: item.ContentType,
				Data:        fp,
			}

			errCh := make(chan error, 1)
			imgOut, err := jpeg.Converter.Convert(img, errCh)
			if err != nil {
				t.Fatal(err)
			}

			_, typ, err := image.Decode(imgOut.Data)
			if err != nil {
				t.Fatal(err)
			}
			if typ != "jpeg" {
				t.Errorf("expected type of converted Convertors to be jpeg, got %s", typ)
			}
			if err := <-errCh; err != nil {
				t.Errorf("got error converting Convertors %s", err)
			}
		})
	}
}

func BenchmarkTransformToJPEG(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, item := range fileTests {
			b.Run(item.Name, func(b *testing.B) {
				fp, err := os.Open(item.Path)
				if err != nil {
					b.Fatal(err)
				}
				defer fp.Close()

				img := Service.Image{
					ID:          item.Name,
					ContentType: item.ContentType,
					Data:        fp,
				}

				errCh := make(chan error, 1)
				imgOut, err := jpeg.Converter.Convert(img, errCh)
				if err != nil {
					b.Fatal(err)
				}

				_, typ, err := image.Decode(imgOut.Data)
				if err != nil {
					b.Fatal(err)
				}
				if typ != "jpeg" {
					b.Errorf("expected type of converted Convertors to be jpeg, got %s", typ)
				}
				if err := <-errCh; err != nil {
					b.Errorf("got error converting Convertors %s", err)
				}
			})
		}
	}
}
