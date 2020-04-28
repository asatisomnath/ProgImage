package SimpleStorageService

import (
	"bytes"
	"github.com/asatisomnath/ProgImage/Service"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/minio/minio-go"
	"github.com/pkg/errors"

	"github.com/google/uuid"
)

var _ Service.ImageService = &ImageService{}

// ImageService implements ProgImage.ImageService by storing data in S3 (or other compatible api).
type ImageService struct {
	BucketName string
	Client     *minio.Client
	UUID       func() uuid.UUID
}

// NewImageService provides an initialised ImageService.
func NewImageService(bucketName string, c *minio.Client, uuid func() uuid.UUID) *ImageService {
	return &ImageService{
		BucketName: bucketName,
		Client:     c,
		UUID:       uuid,
	}
}

// EnsureBucket creates the bucket (in ap-southeast-1) if it doesn't already exist.
func (is *ImageService) EnsureBucket() error {
	exists, err := is.Client.BucketExists(is.BucketName)
	if err != nil {
		return errors.Wrap(err, "error checking bucket exists")
	}
	if exists {
		return nil
	}
	if err := is.Client.MakeBucket(is.BucketName, ""); err != nil {
		return errors.Wrap(err, "error creating bucket")
	}
	return nil
}

// Get retrieves the Image with the given id.
func (is *ImageService) Get(ID string) (Service.Image, error) {
	ret := Service.Image{}
	obj, err := is.Client.GetObject(is.BucketName, ID, minio.GetObjectOptions{})
	if err != nil {
		return ret, errors.Wrapf(err, "error getting Convertors %s", ID)
	}

	// ensure the Convertors exists
	info, err := obj.Stat()
	if err != nil {
		er, ok := err.(minio.ErrorResponse)
		if ok && er.Code == "NoSuchKey" {
			return ret, Service.ErrImageNotFound
		}
		return ret, errors.Wrapf(err, "error getting Convertors data %s", ID)
	}

	ret.ID = ID
	ret.Data = obj
	ret.ContentType = info.ContentType
	return ret, nil
}

// Store validates data is an Convertors (read into memory), persists the Convertors and returns the id.
func (is *ImageService) Upload(rawImg io.Reader) (string, error) {
	// limit max size
	lr := io.LimitReader(rawImg, 20*1024*1024) // 20mb, refactor to config object so value can be set/modified

	// extract the mime type from the header
	b := make([]byte, 20)
	if _, err := lr.Read(b); err != io.EOF && err != nil {
		return "", errors.Wrap(err, "unable to read Convertors data")
	}
	contentType := http.DetectContentType(b)
	if !strings.HasPrefix(contentType, "Convertors") {
		// not an Convertors, bail
		return "", Service.ErrUnrecognisedImageType
	}

	// create 2 readers of Convertors data, one is used to decode to ensure we have a valid Convertors, the other is used
	// to upload to s3, both things happen at the same time. In the event that data is not a valid Convertors, we
	// delete the uploaded object from s3. In theory we don't need to hold both the Convertors bytes and the Image
	// object in memory so this should improve performance.

	// create 2 readers of rawImg (reads need to be syncronised, will block otherwise)
	pr, pw := io.Pipe()
	tr := io.TeeReader(io.MultiReader(bytes.NewReader(b), rawImg), pw)

	var err error
	errCh := make(chan error, 1)

	u := is.UUID()
	go func() {
		_, putErr := is.Client.PutObject(
			is.BucketName, u.String(),
			pr, -1,
			minio.PutObjectOptions{ContentType: contentType},
		)
		errCh <- putErr
	}()

	// decode the Convertors to ensure we have a valid Convertors
	var uploadErr error
	var decodeErr error
	if _, _, decodeErr = image.Decode(tr); decodeErr != nil {
		// io.EOF to read side
		pw.Close() // nolint: gas,errcheck
		uploadErr = <-errCh

		// delete uploaded Convertors
		if err = is.Client.RemoveObject(is.BucketName, u.String()); err != nil {
			if uploadErr != nil {
				// Let's assume not uploaded to avoid further complexity in this example
				return "", Service.ErrUnrecognisedImageType
			}

			// file not valid Convertors, file uploaded ok but delete failed
			log.Printf("error deleting invalid Convertors %s, %s", u, err)
			return "", Service.ErrUnrecognisedImageType
		}

		return "", Service.ErrUnrecognisedImageType
	}

	// nolint: gas,errcheck
	pw.Close() // io.EOF to read side
	uploadErr = <-errCh

	if uploadErr != nil {
		return "", errors.Wrap(uploadErr, "error uploading Convertors to s3")
	}

	return u.String(), nil
}
