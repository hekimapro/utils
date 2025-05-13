package file

import (
	"bytes"
	"image"
	"io"
	"log"
	"strings"

	_ "image/gif"  // Register the GIF format
	_ "image/jpeg" // Register the JPEG format
	_ "image/png"  // Register the PNG format

	"github.com/chai2010/webp"
)

// ConvertToWebP converts an image file to WebP format
// Decodes the input image and encodes it as lossless WebP
func convertToWebP(file io.Reader) (io.Reader, error) {
	// Decode the input image to an image.Image
	img, _, err := image.Decode(file)
	if err != nil {
		// Log and return error if image decoding fails
		log.Printf("Error decoding image: %v", err)
		return nil, err
	}

	// Initialize a buffer to store the WebP-encoded image
	var webpBuffer bytes.Buffer

	// Encode the image to WebP format with lossless compression
	err = webp.Encode(&webpBuffer, img, &webp.Options{Lossless: true})
	if err != nil {
		// Log and return error if WebP encoding fails
		log.Printf("Error encoding image to WebP: %v", err)
		return nil, err
	}

	// Return the WebP image as an io.Reader
	return &webpBuffer, nil
}

// CheckAndConvertFile checks if a file is an image and converts it to WebP
// Supports JPG, JPEG, and PNG formats; returns original file for unsupported types
func CheckAndConvertFile(file io.Reader, fileName string) (io.Reader, error) {
	// Extract the file extension and convert to lowercase
	ext := strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])
	// Check if the file is a supported image format
	if ext != "jpg" && ext != "jpeg" && ext != "png" {
		// Log and return the original file for unsupported formats
		log.Println("File format not supported, returning original file.")
		return file, nil
	}

	// Convert the supported image to WebP format
	convertedFile, err := convertToWebP(file)
	if err != nil {
		// Log and return error if conversion fails
		log.Printf("Error converting file: %v", err)
		return nil, err
	}

	// Return the converted WebP file
	return convertedFile, nil
}