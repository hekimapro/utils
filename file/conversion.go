package file

import (
	"bytes"   // bytes provides utilities for byte buffer manipulation.
	"image"   // image provides image decoding and format registration.
	"io"      // io provides interfaces for I/O operations.
	"strings" // strings provides utilities for string manipulation.

	_ "image/gif"  // Register GIF format for image decoding.
	_ "image/jpeg" // Register JPEG format for image decoding.
	_ "image/png"  // Register PNG format for image decoding.

	"github.com/chai2010/webp"       // webp provides WebP encoding and decoding.
	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
)

// convertToWebP converts an image file to WebP format.
// Returns the converted image as an io.Reader or an error if conversion fails.
func convertToWebP(file io.Reader) (io.Reader, error) {
	// Log the start of the image decoding process.
	log.Info("🖼️ Decoding input image...")

	// Decode the input image to a generic image.Image type.
	img, _, err := image.Decode(file)
	if err != nil {
		// Log and return an error if image decoding fails.
		log.Error("❌ Failed to decode image: " + err.Error())
		return nil, err
	}

	// Create a buffer to store the WebP-encoded image.
	var webpBuffer bytes.Buffer

	// Encode the image to WebP format using lossless compression.
	log.Info("🧪 Encoding image to WebP format (lossless)")
	err = webp.Encode(&webpBuffer, img, &webp.Options{Lossless: true})
	if err != nil {
		// Log and return an error if WebP encoding fails.
		log.Error("❌ Failed to encode image to WebP: " + err.Error())
		return nil, err
	}

	// Log successful conversion to WebP.
	log.Success("✅ Image successfully converted to WebP format")
	// Return the WebP image as an io.Reader.
	return &webpBuffer, nil
}

// CheckAndConvertFile checks if a file is an image and converts it to WebP.
// Returns the original file if the format is unsupported, or the WebP-converted file.
// Returns an error if conversion fails.
func CheckAndConvertFile(file io.Reader, fileName string) (io.Reader, error) {
	// Log the start of the file type checking process.
	log.Info("🔍 Checking file type for WebP conversion")

	// Extract the file extension (case-insensitive) from the file name.
	ext := strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])
	// Check if the file extension is a supported image format (jpg, jpeg, png).
	if ext != "jpg" && ext != "jpeg" && ext != "png" {
		// Log and return the original file if the format is unsupported.
		log.Info("ℹ️ Unsupported image format '" + ext + "'. Skipping WebP conversion.")
		return file, nil
	}

	// Log that a supported image format was detected.
	log.Info("🟢 Supported image format detected (" + ext + "). Proceeding with WebP conversion")
	// Convert the image to WebP format.
	convertedFile, err := convertToWebP(file)
	if err != nil {
		// Log and return an error if WebP conversion fails.
		log.Error("❌ WebP conversion failed: " + err.Error())
		return nil, err
	}

	// Log successful WebP conversion.
	log.Success("🎉 File successfully converted to WebP format")
	return convertedFile, nil
}