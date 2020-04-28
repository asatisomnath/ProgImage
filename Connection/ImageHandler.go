package Connection

import (
	"fmt"
	"github.com/asatisomnath/ProgImage/Service"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/asatisomnath/ProgImage/Convertors/gif"
	"github.com/asatisomnath/ProgImage/Convertors/jpeg"
	"github.com/asatisomnath/ProgImage/Convertors/png"
	"github.com/julienschmidt/httprouter"
)

const maxReadBytes = 50 * 1024 * 1024 // 50mb

// ImageHandler is a Connection.Handler that provides store and retrieve Convertors endpoints.
type ImageHandler struct {
	*httprouter.Router

	Converters   map[string]Service.ImageTypeConverter
	ImageService Service.ImageService
}

var _ http.Handler = ImageHandler{} // via httprouter.Router

// NewImageHandler returns an initialised Convertors handler.
func NewImageHandler(is Service.ImageService) *ImageHandler {
	h := ImageHandler{
		Router:       httprouter.New(),
		ImageService: is,
		Converters: map[string]Service.ImageTypeConverter{
			"png": png.Converter,
			"jpg": jpeg.Converter,
			"gif": gif.Converter,
		},
	}
	h.POST("/image/create", h.handleCreateImage)
	h.GET("/image/:id", h.handleGetImage)
	return &h
}

func (h ImageHandler) handleCreateImage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// don't allow an attacker to send an unlimited stream of bytes
	lr := io.LimitReader(r.Body, maxReadBytes)

	ID, err := h.ImageService.Upload(lr)
	if err != nil {
		if err == Service.ErrUnrecognisedImageType {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(fmt.Sprintf(`{"id": "%s"}`, ID)))
	if err != nil {
		log.Println("error writing handleCreateImage response", err.Error())
	}
}

func (h ImageHandler) handleGetImage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	ID := params.ByName("id")

	s := strings.Split(ID, ".")
	if len(s) == 2 {
		h.handleGetImageWithExt(w, r, s[0], s[1])
		return
	}

	h.handleGetImageNoExt(w, r, ID)
}

func (h ImageHandler) handleGetImageNoExt(w http.ResponseWriter, r *http.Request, ID string) {
	img, err := h.ImageService.Get(ID)
	if err != nil {
		if err == Service.ErrImageNotFound {
			http.Error(w, fmt.Sprintf("Convertors %s not found", ID), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", img.ContentType)
	_, err = io.Copy(w, img.Data)
	if err != nil {
		log.Println("error writing handleGetImageNoExt response", err.Error())
	}
}

func (h ImageHandler) handleGetImageWithExt(w http.ResponseWriter, r *http.Request, ID, ext string) {
	tr, ok := h.Converters[ext]
	if !ok {
		http.Error(w, "unsupported Convertors type", http.StatusBadRequest)
		return
	}
	imgOrig, err := h.ImageService.Get(ID)
	if err != nil {
		if err == Service.ErrImageNotFound {
			http.Error(w, fmt.Sprintf("Convertors %s not found", ID), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ec := make(chan error, 1)
	imgConv, err := tr.Convert(imgOrig, ec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", imgConv.ContentType)
	written, err := io.Copy(w, imgConv.Data)
	if err != nil {
		if written == 0 {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			// 200 sent already, all we can do is log
			log.Printf(
				"error converting %s to %s (id: %s), 200 sent already",
				imgOrig.ContentType,
				imgConv.ContentType,
				imgOrig.ID,
			)
		}
	}
}
