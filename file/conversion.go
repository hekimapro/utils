package file

import (
	"bytes"
	"image"
	"io"
	"strings"

	_ "image/gif"  // Register the GIF format
	_ "image/jpeg" // Register the JPEG format
	_ "image/png"  // Register the PNG format

	"github.com/chai2010/webp"
	"github.com/hekimapro/utils/log"
)

// convertToWebP converts an image file to WebP format
func convertToWebP(file io.Reader) (io.Reader, error) {
	log.Info("Decoding input image")

	img, _, err := image.Decode(file)
	if err != nil {
		log.Error("Failed to decode image: " + err.Error())
		return nil, err
	}

	var webpBuffer bytes.Buffer

	log.Info("Encoding image to WebP format")
	err = webp.Encode(&webpBuffer, img, &webp.Options{Lossless: true})
	if err != nil {
		log.Error("Failed to encode image to WebP: " + err.Error())
		return nil, err
	}

	log.Success("Image successfully converted to WebP")
	return &webpBuffer, nil
}

// CheckAndConvertFile checks if a file is an image and converts it to WebP
func CheckAndConvertFile(file io.Reader, fileName string) (io.Reader, error) {
	log.Info("Checking file type for conversion")

	ext := strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])
	if ext != "jpg" && ext != "jpeg" && ext != "png" {
		log.Info("Unsupported image format: " + ext + ". Skipping conversion.")
		return file, nil
	}

	log.Info("Image format supported (" + ext + "). Converting to WebP")
	convertedFile, err := convertToWebP(file)
	if err != nil {
		log.Error("Image conversion to WebP failed: " + err.Error())
		return nil, err
	}

	log.Success("File successfully converted to WebP format")
	return convertedFile, nil
}
