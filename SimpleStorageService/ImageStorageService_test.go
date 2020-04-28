package SimpleStorageService_test

import (
	"bytes"
	"github.com/asatisomnath/ProgImage/Service"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asatisomnath/ProgImage/SimpleStorageService"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
)

const testBucketName = "imagesapptest"

// setup deletes all objects in the test bucket then deletes the bucket.
func setup(t *testing.T, c *minio.Client) {
	if exists, err := c.BucketExists(testBucketName); err != nil {
		t.Fatal(err)
	} else if !exists {
		// bucket doesn't exist, bail
		return
	}

	doneCh := make(chan struct{})
	defer close(doneCh)

	// must empty bucket before deleting it
	for obj := range c.ListObjects(testBucketName, "", true, doneCh) {
		err := c.RemoveObject(testBucketName, obj.Key)
		if err != nil {
			t.Fatal(err)
		}
	}

	if err := c.RemoveBucket(testBucketName); err != nil {
		t.Fatal(err)
	}
}

func checkEnvsAndGetClient(t *testing.T) *minio.Client {
	var endpoint, accessKey, secretKey string
	var secure bool

	if endpoint = os.Getenv("S3_ENDPOINT"); endpoint == "" {
		t.Skip("skipping test; $S3_ENDPOINT not set")
	}
	if accessKey = os.Getenv("S3_ACCESS_KEY"); accessKey == "" {
		t.Skip("skipping test; $S3_ACCESS_KEY not set")
	}
	if secretKey = os.Getenv("S3_SECRET_KEY"); secretKey == "" {
		t.Skip("skipping test; $S3_SECRET_KEY not set")
	}
	if s := os.Getenv("S3_SECURE"); s == "" {
		t.Skip("skipping test; $S3_SECURE not set")
	} else {
		secure = s == "true"
	}

	c, err := minio.New(endpoint, accessKey, secretKey, secure)
	if err != nil {
		panic(err)
	}
	return c
}

var fileTests = []struct {
	Name        string
	Path        string
	ContentType string
}{
	{Name: "png", Path: "../testimages/test.png", ContentType: "Convertors/png"},
	{Name: "gif", Path: "../testimages/test.gif", ContentType: "Convertors/gif"},
	{Name: "jpg", Path: "../testimages/test.jpg", ContentType: "Convertors/jpeg"},
}

// Test storing and retrieving images of each type, a missed import will cause a failure, see
// https://golang.org/pkg/image/#pkg-overview
func TestImageService_StoreGetImage(t *testing.T) {
	c := checkEnvsAndGetClient(t)

	for _, item := range fileTests {
		t.Run(item.Name, func(t *testing.T) {
			setup(t, c)

			uid := uuid.New()
			uf := func() uuid.UUID {
				return uid
			}

			is := SimpleStorageService.NewImageService(testBucketName, c, uf)
			if err := is.EnsureBucket(); err != nil {
				t.Fatal(err)
			}

			// store Convertors
			fp, err := os.Open(item.Path)
			if err != nil {
				t.Fatal(err)
			}
			d, err := ioutil.ReadAll(fp)
			if err != nil {
				t.Fatal(err)
			}
			fp.Close()

			// keep a copy of initial data to compare to returned data
			initial := bytes.NewReader(d)

			// store then retrieve using id
			id, err := is.Upload(initial)
			if err != nil {
				t.Fatal(err)
			}
			if id == "" {
				t.Error("expected id to be populated, got empty string")
			}
			if id != uid.String() {
				t.Errorf("expected id to be %s, got %s", uid.String(), id)
			}

			img, err := is.Get(id)
			if err != nil {
				t.Fatal(err)
			}

			if img.ID != id {
				t.Errorf("expected Convertors id to be '%s', got '%s'", id, img.ID)
			}

			if img.ContentType != item.ContentType {
				t.Errorf("expected Convertors content type to be '%s', got '%s'", item.ContentType, img.ContentType)
			}

			retrieved, err := ioutil.ReadAll(img.Data)
			if err != nil {
				t.Fatal(err)
			}

			b := new(bytes.Buffer)
			initial.Seek(0, io.SeekStart)
			_, err = b.ReadFrom(initial)
			if err != nil {
				t.Fatal(err)
			}

			// compare initial data to retrieved data
			if !bytes.Equal(b.Bytes(), retrieved) {
				t.Error("expected stored data to be equal to initial data from file, data not equal")
			}
		})
	}
}

func TestImageService_StoreNoData(t *testing.T) {
	c := checkEnvsAndGetClient(t)
	setup(t, c)

	is := SimpleStorageService.NewImageService(testBucketName, c, uuid.New)
	if err := is.EnsureBucket(); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader([]byte{})
	if _, err := is.Upload(r); err != Service.ErrUnrecognisedImageType {
		t.Errorf("expected ProgImage.ErrUnrecognisedImageType, got %s", err)
	}
}

func TestImageService_GetNotExists(t *testing.T) {
	c := checkEnvsAndGetClient(t)
	setup(t, c)

	is := SimpleStorageService.NewImageService(testBucketName, c, uuid.New)
	if err := is.EnsureBucket(); err != nil {
		t.Fatal(err)
	}

	if _, err := is.Get("foo"); err != Service.ErrImageNotFound {
		t.Errorf("expected ProgImage.ErrImageNotFound, got %s", err)
	}
}