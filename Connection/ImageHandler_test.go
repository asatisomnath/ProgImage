package Connection_test

import (
	"bytes"
	"encoding/json"
	"github.com/asatisomnath/ProgImage/Service"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	pihttp "github.com/asatisomnath/ProgImage/Connection"
	"github.com/asatisomnath/ProgImage/Mock"
)

// ImageHandler is test wrapper that uses a mocked image service.
type ImageHandler struct {
	*pihttp.ImageHandler

	ImageService *Mock.ImageService
}

func NewImageHandler() *ImageHandler {
	is := new(Mock.ImageService)
	ih := pihttp.NewImageHandler(is)
	return &ImageHandler{ih, is}
}

func TestGet_OK(t *testing.T) {
	h := NewImageHandler()

	var createdID string
	h.ImageService.GetFunc = func(ID string) (Service.Image, error) {
		createdID = ID
		return Service.Image{Data: new(bytes.Reader)}, nil
	}

	expectedID := "foo"
	req, err := http.NewRequest("GET", "/image/"+expectedID, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected: %v got: %v", http.StatusOK, status)
	}

	if expectedID != createdID {
		t.Errorf("expected id to be: %v got: %v", expectedID, createdID)
	}
}

func TestGet_NotFound(t *testing.T) {
	h := NewImageHandler()

	h.ImageService.GetFunc = func(ID string) (Service.Image, error) {
		return Service.Image{}, Service.ErrImageNotFound
	}

	expectedID := "foo"
	req, err := http.NewRequest("GET", "/image/"+expectedID, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("expected: %v got: %v", http.StatusNotFound, status)
	}
}

func TestGet_WithExt(t *testing.T) {
	h := NewImageHandler()

	var getID string
	var fp *os.File
	h.ImageService.GetFunc = func(ID string) (Service.Image, error) {
		getID = ID

		var err error
		fp, err = os.Open("../testimages/test.jpg")
		if err != nil {
			return Service.Image{}, err
		}

		return Service.Image{ID: ID, Data: fp, ContentType: "image/jpeg"}, nil
	}

	expectedID := "foo.png"
	req, err := http.NewRequest("GET", "/image/"+expectedID, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected: %v got: %v", http.StatusOK, status)
	}

	if rr.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected Content-Type image/png, got: %v", rr.Header().Get("Content-Type"))
	}

	if getID != strings.Split(expectedID, ".")[0] {
		t.Errorf("expected get id to be: %v got: %v", expectedID, getID)
	}

	fp.Close()
}

func TestGet_WithUnsupportedExt(t *testing.T) {
	h := NewImageHandler()

	h.ImageService.GetFunc = func(ID string) (Service.Image, error) {
		return Service.Image{ID: ID, Data: new(bytes.Buffer), ContentType: "image/jpeg"}, nil
	}

	expectedID := "foo.unsupported"
	req, err := http.NewRequest("GET", "/image/"+expectedID, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected: %v got: %v", http.StatusOK, status)
	}
}

func TestStore_OK(t *testing.T) {
	h := NewImageHandler()

	expectedID := "foo"
	dataIn := new(bytes.Buffer)
	h.ImageService.StoreFunc = func(r io.Reader) (string, error) {
		io.Copy(dataIn, r)
		return expectedID, nil
	}

	imgData := "some img data"
	req, err := http.NewRequest("POST", "/image/create", bytes.NewReader([]byte(imgData)))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("expected: %v got: %v", http.StatusCreated, status)
	}

	type respData struct {
		ID string
	}
	rd := new(respData)
	if err := json.NewDecoder(rr.Body).Decode(rd); err != nil {
		t.Fatal(err)
	}

	if expectedID != rd.ID {
		t.Errorf("expected id to be: %v got: %v", expectedID, rd)
	}

	if imgData != dataIn.String() {
		t.Errorf("expected data read to be '%s', got '%s'", imgData, dataIn.String())
	}
}

func TestStore_UnregognisedImagetype(t *testing.T) {
	h := NewImageHandler()

	h.ImageService.StoreFunc = func(r io.Reader) (string, error) {
		return "", Service.ErrUnrecognisedImageType
	}

	req, err := http.NewRequest("POST", "/image/create", new(bytes.Reader))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected: %v got: %v", http.StatusBadRequest, status)
	}
}
