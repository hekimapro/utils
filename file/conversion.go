package file

import (
	"bytes"   // bytes provides utilities for byte buffer manipulation.
	"context" // context provides support for cancellation and timeouts.
	"fmt"
	"image"   // image provides image decoding and format registration.
	"io"      // io provides interfaces for I/O operations.
	"strings" // strings provides utilities for string manipulation.
	"time"    // time provides functionality for timeouts and durations.

	_ "image/gif"  // Register GIF format for image decoding.
	_ "image/jpeg" // Register JPEG format for image decoding.
	_ "image/png"  // Register PNG format for image decoding.

	"github.com/chai2010/webp"           // webp provides WebP encoding and decoding.
	"github.com/hekimapro/utils/helpers" // helpers provides utility functions.
	"github.com/hekimapro/utils/log"     // log provides colored logging utilities.
)

// convertToWebP converts an image file to WebP format with context support.
// Returns the converted image as an io.Reader or an error if conversion fails.
func convertToWebP(ctx context.Context, file io.Reader) (io.Reader, error) {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "WebP conversion cancelled before start")
	default:
		// Continue with conversion
	}

	// Log the start of the image decoding process.
	log.Info("🖼️ Decoding input image...")

	// Decode the input image to a generic image.Image type.
	img, _, err := image.Decode(file)
	if err != nil {
		// Log and return an error if image decoding fails.
		log.Error("❌ Failed to decode image: " + err.Error())
		return nil, helpers.WrapError(err, "failed to decode image")
	}

	// Check context cancellation after decoding
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "WebP conversion cancelled after decoding")
	default:
		// Continue with encoding
	}

	// Create a buffer to store the WebP-encoded image.
	var webpBuffer bytes.Buffer

	// Encode the image to WebP format using lossless compression.
	log.Info("🧪 Encoding image to WebP format (lossless)")
	err = webp.Encode(&webpBuffer, img, &webp.Options{Lossless: true})
	if err != nil {
		// Log and return an error if WebP encoding fails.
		log.Error("❌ Failed to encode image to WebP: " + err.Error())
		return nil, helpers.WrapError(err, "failed to encode image to WebP")
	}

	// Check context cancellation after encoding
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "WebP conversion cancelled after encoding")
	default:
		// Continue with success
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
	// Create context with timeout for conversion operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

	// Check context cancellation before starting conversion
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "file conversion cancelled before start")
	default:
		// Continue with conversion
	}

	// Log that a supported image format was detected.
	log.Info("🟢 Supported image format detected (" + ext + "). Proceeding with WebP conversion")

	// Convert the image to WebP format with context support.
	convertedFile, err := convertToWebP(ctx, file)
	if err != nil {
		// Log and return an error if WebP conversion fails.
		log.Error("❌ WebP conversion failed: " + err.Error())
		return nil, helpers.WrapError(err, "WebP conversion failed")
	}

	// Log successful WebP conversion.
	log.Success("🎉 File successfully converted to WebP format")
	return convertedFile, nil
}

// CheckAndConvertFileWithOptions checks if a file is an image and converts it to WebP with custom options.
// Returns the original file if the format is unsupported, or the WebP-converted file.
// Returns an error if conversion fails.
func CheckAndConvertFileWithOptions(file io.Reader, fileName string, lossless bool, quality float32) (io.Reader, error) {
	// Create context with timeout for conversion operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log the start of the file type checking process.
	log.Info("🔍 Checking file type for WebP conversion with custom options")

	// Extract the file extension (case-insensitive) from the file name.
	ext := strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])
	// Check if the file extension is a supported image format (jpg, jpeg, png).
	if ext != "jpg" && ext != "jpeg" && ext != "png" && ext != "gif" {
		// Log and return the original file if the format is unsupported.
		log.Info("ℹ️ Unsupported image format '" + ext + "'. Skipping WebP conversion.")
		return file, nil
	}

	// Check context cancellation before starting conversion
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "file conversion cancelled before start")
	default:
		// Continue with conversion
	}

	// Log that a supported image format was detected.
	log.Info("🟢 Supported image format detected (" + ext + "). Proceeding with WebP conversion")

	// Convert the image to WebP format with custom options and context support.
	convertedFile, err := convertToWebPWithOptions(ctx, file, lossless, quality)
	if err != nil {
		// Log and return an error if WebP conversion fails.
		log.Error("❌ WebP conversion with options failed: " + err.Error())
		return nil, helpers.WrapError(err, "WebP conversion with options failed")
	}

	// Log successful WebP conversion.
	log.Success("🎉 File successfully converted to WebP format with custom options")
	return convertedFile, nil
}

// convertToWebPWithOptions converts an image file to WebP format with custom options and context support.
func convertToWebPWithOptions(ctx context.Context, file io.Reader, lossless bool, quality float32) (io.Reader, error) {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "WebP conversion with options cancelled before start")
	default:
		// Continue with conversion
	}

	// Log the start of the image decoding process.
	log.Info("🖼️ Decoding input image for custom WebP conversion...")

	// Decode the input image to a generic image.Image type.
	img, _, err := image.Decode(file)
	if err != nil {
		// Log and return an error if image decoding fails.
		log.Error("❌ Failed to decode image: " + err.Error())
		return nil, helpers.WrapError(err, "failed to decode image")
	}

	// Check context cancellation after decoding
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "WebP conversion with options cancelled after decoding")
	default:
		// Continue with encoding
	}

	// Create a buffer to store the WebP-encoded image.
	var webpBuffer bytes.Buffer

	// Log the conversion options
	if lossless {
		log.Info("🧪 Encoding image to WebP format (lossless)")
	} else {
		log.Info("🧪 Encoding image to WebP format (lossy, quality: " + fmt.Sprintf("%.1f", quality) + ")")
	}

	// Encode the image to WebP format with custom options.
	err = webp.Encode(&webpBuffer, img, &webp.Options{
		Lossless: lossless,
		Quality:  quality,
	})
	if err != nil {
		// Log and return an error if WebP encoding fails.
		log.Error("❌ Failed to encode image to WebP with options: " + err.Error())
		return nil, helpers.WrapError(err, "failed to encode image to WebP with options")
	}

	// Check context cancellation after encoding
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "WebP conversion with options cancelled after encoding")
	default:
		// Continue with success
	}

	// Log successful conversion to WebP.
	log.Success("✅ Image successfully converted to WebP format with custom options")
	// Return the WebP image as an io.Reader.
	return &webpBuffer, nil
}

// IsImageFormatSupported checks if the file format is supported for WebP conversion.
func IsImageFormatSupported(fileName string) bool {
	ext := strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])
	supportedFormats := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
	}
	return supportedFormats[ext]
}

// GetSupportedImageFormats returns a list of supported image formats for WebP conversion.
func GetSupportedImageFormats() []string {
	return []string{"jpg", "jpeg", "png", "gif"}
}

// GetImageInfo attempts to get basic information about the image.
func GetImageInfo(file io.Reader) (string, image.Config, error) {
	// Create context with timeout for image info operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return "", image.Config{}, helpers.WrapError(ctx.Err(), "image info operation cancelled")
	default:
		// Continue with operation
	}

	// Decode the image config to get dimensions without decoding the entire image
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return "", image.Config{}, helpers.WrapError(err, "failed to decode image config")
	}

	// Check context cancellation after decoding
	select {
	case <-ctx.Done():
		return "", image.Config{}, helpers.WrapError(ctx.Err(), "image info operation cancelled after decoding")
	default:
		// Continue with success
	}

	return format, config, nil
}
