package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"

	"github.com/nfnt/resize"
)

func downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}

func zoomAndCrop(img image.Image, zoom float64) image.Image {
	newSize := uint(float64(img.Bounds().Dx()) * zoom)
	newImg := resize.Resize(newSize, 0, img, resize.Lanczos3)

	deltaX := (newImg.Bounds().Dx() - img.Bounds().Dx()) / 2
	deltaY := (newImg.Bounds().Dy() - img.Bounds().Dy()) / 2

	croppedRect := image.Rect(deltaX, deltaY, newImg.Bounds().Dx()-deltaX, newImg.Bounds().Dy()-deltaY)
	return newImg.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(croppedRect)
}

func handler(w http.ResponseWriter, r *http.Request) {
	encURL := r.URL.Query().Get("url")
	zoomStr := r.URL.Query().Get("zoom")

	if encURL == "" || zoomStr == "" {
		http.Error(w, "url and zoom parameters are required", http.StatusBadRequest)
		return
	}

	url, err := base64.StdEncoding.DecodeString(encURL)
	if err != nil {
		http.Error(w, "Invalid base64 encoded url", http.StatusBadRequest)
		return
	}

	zoom, err := strconv.ParseFloat(zoomStr, 64)
	if err != nil || zoom <= 0 {
		http.Error(w, "Invalid zoom value", http.StatusBadRequest)
		return
	}

	fmt.Printf("Downloading %s with zoom %f\n", url, zoom)

	img, err := downloadImage(string(url))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to download image", http.StatusInternalServerError)
		return
	}

	zoomedImg := zoomAndCrop(img, zoom)

	w.Header().Set("Content-Type", "image/jpeg")
	jpeg.Encode(w, zoomedImg, nil)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}
