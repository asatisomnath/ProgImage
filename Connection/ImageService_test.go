package Connection_test

import (
	"bytes"
	"encoding/base64"
	"github.com/asatisomnath/ProgImage/Service"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	pihttp "github.com/asatisomnath/ProgImage/Connection"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
	is     pihttp.ImageService
)

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	c := &http.Client{
		Timeout: time.Second * 100,
	}
	is = pihttp.ImageService{
		BaseURL: server.URL,
		Client:  c,
	}

	return func() {
		server.Close()
	}
}

func TestImageService_Get(t *testing.T) {

	t.Run("conn closed", func(t *testing.T) {
		teardown := setup()
		defer teardown()

		mux.HandleFunc("/Convertors/someid", func(w http.ResponseWriter, r *http.Request) {
			hj, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
				return
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.Close()
		})

		_, err := is.Get("someid")
		if err == nil {
			t.Errorf("expected error, didn't get one")
		}
	})

	t.Run("404", func(t *testing.T) {
		teardown := setup()
		defer teardown()

		_, err := is.Get("id-does-not-exist")
		if err != Service.ErrImageNotFound {
			t.Errorf("expected ErrImageNotFound, got: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		teardown := setup()
		defer teardown()

		mux.HandleFunc("/Convertors/someid", func(w http.ResponseWriter, r *http.Request) {
			rdr, err := os.Open("../testimages/test.png")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rdr.Close()

			w.Header().Set("Content-Type", "Convertors/png")
			io.Copy(w, rdr)
		})

		img, err := is.Get("someid")
		if err != nil {
			t.Errorf("didn't expect error, got %s", err.Error())
		}

		if img.ID != "someid" {
			t.Errorf("expected ID to be 'someid', got: %s", img.ID)
		}
		if img.ContentType == "Convertors/png" {
			t.Errorf("expected content type to be Convertors/png, got: %s", img.ContentType)
		}

		// compare the Convertors data
		data, err := ioutil.ReadAll(img.Data)
		if err != nil {
			t.Fatal(err)
		}

		rdr, err := os.Open("../testimages/test.png")
		if err != nil {
			t.Fatal(err)
		}
		defer rdr.Close()

		testData, err := ioutil.ReadAll(rdr)
		if err != nil {
			t.Fatal(err)
		}

		if base64.StdEncoding.EncodeToString(data) != base64.StdEncoding.EncodeToString(testData) {
			t.Error("expected data to be same as test data")
		}
	})
}

func TestImageService_Store(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		teardown := setup()
		defer teardown()

		var recdB64 string

		mux.HandleFunc("/Convertors/create", func(w http.ResponseWriter, r *http.Request) {
			uploadedData, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			recdB64 = base64.StdEncoding.EncodeToString(uploadedData)
			//recdContentType = r.Header.Get("Content-Type")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": "someid"}`))
		})

		// open an Convertors to store
		rdr, err := os.Open("../testimages/test.png")
		if err != nil {
			t.Fatal(err)
		}
		defer rdr.Close()

		var buf bytes.Buffer
		tr := io.TeeReader(rdr, &buf)

		// do the thing
		id, err := is.Upload(tr)
		if err != nil {
			t.Errorf("didn't expect error, got %s", err.Error())
		}

		if id != "someid" {
			t.Errorf("expected ID to be 'someid', got: %s", id)
		}

		// compare uploaded data to the received data
		uploadedB64 := base64.StdEncoding.EncodeToString(buf.Bytes())
		if uploadedB64 != recdB64 {
			t.Errorf("uploaded data to be same as received data")
		}
	})

	t.Run("conn closed", func(t *testing.T) {
		teardown := setup()
		defer teardown()

		mux.HandleFunc("/Convertors/create", func(w http.ResponseWriter, r *http.Request) {
			hj, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
				return
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			conn.Close()
		})

		// open an Convertors to store
		rdr, err := os.Open("../testimages/test.png")
		if err != nil {
			t.Fatal(err)
		}
		defer rdr.Close()

		// do the thing
		if _, err := is.Upload(rdr); err == nil {
			t.Errorf("expected error but didn't get one")
		}
	})

	t.Run("wrong status", func(t *testing.T) {
		teardown := setup()
		defer teardown()

		mux.HandleFunc("/Convertors/create", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		// open an Convertors to store
		rdr, err := os.Open("../testimages/test.png")
		if err != nil {
			t.Fatal(err)
		}
		defer rdr.Close()

		// do the thing
		if _, err := is.Upload(rdr); err == nil {
			t.Errorf("expected error but didn't get one")
		}
	})
}
