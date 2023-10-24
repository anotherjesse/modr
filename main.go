package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"

	imagick "gopkg.in/gographics/imagick.v3/imagick"
)

func zoomImage(mw *imagick.MagickWand, zoom float64) error {

	// Perform the zoom and crop
	width := mw.GetImageWidth()
	height := mw.GetImageHeight()
	newWidth := uint(float64(width) * zoom)
	newHeight := uint(float64(height) * zoom)

	// Resize
	mw.ResizeImage(newWidth, newHeight, imagick.FILTER_LANCZOS)

	// Crop
	x := (newWidth - width) / 2
	y := (newHeight - height) / 2
	return mw.CropImage(width, height, int(x), int(y))
}

func handler(w http.ResponseWriter, r *http.Request) {
	encURL := r.URL.Query().Get("url")
	zoomStr := r.URL.Query().Get("zoom")

	if encURL == "" || zoomStr == "" {
		http.Error(w, "url and zoom parameters are required", http.StatusBadRequest)
		return
	}

	url_bytes, err := base64.StdEncoding.DecodeString(encURL)
	url := string(url_bytes)

	if err != nil {
		http.Error(w, "Invalid base64 encoded url", http.StatusBadRequest)
		return
	}

	fmt.Println("Downloading image from", url)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to download image", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the body data to byte slice
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read image data", http.StatusInternalServerError)
		return
	}

	fmt.Println("Downloaded", len(data), "bytes")

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read the image from the byte slice
	if err := mw.ReadImageBlob(data); err != nil {
		http.Error(w, "Failed to read image data", http.StatusInternalServerError)
		return
	}

	if zoomStr != "" {

		zoom, err := strconv.ParseFloat(zoomStr, 64)
		if err != nil || zoom <= 0 {
			http.Error(w, "Invalid zoom value", http.StatusBadRequest)
			return
		}

		err = zoomImage(mw, zoom)
		if err != nil {
			http.Error(w, "Failed to download or process image", http.StatusInternalServerError)
			return
		}
	}

	content := mw.GetImageBlob()

	if content != nil {
		w.Header().Set("Content-Type", "image/jpeg")
		fmt.Println("Content length:", len(content))
		w.Write(content)
	} else {
		http.Error(w, "No content", http.StatusInternalServerError)
	}
}

func main() {
	imagick.Initialize()
	defer imagick.Terminate()

	http.HandleFunc("/", handler)
	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}
