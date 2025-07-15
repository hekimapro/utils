package file

import (
	"fmt"           // fmt provides formatting and printing functions.
	"io"            // io provides interfaces for I/O operations.
	"os"            // os provides file system operations.
	"path/filepath" // filepath provides utilities for file path manipulation.
	"regexp"        // regexp provides regular expression utilities.
	"strings"       // strings provides utilities for string manipulation.

	"github.com/google/uuid"         // uuid provides UUID generation.
	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
)

// toKebabCase converts a string to kebab-case (lowercase with hyphens).
// Returns the converted string.
func toKebabCase(stringValue string) string {
	// Replace non-alphanumeric characters with hyphens.
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	kebab := re.ReplaceAllString(stringValue, "-")

	// Insert hyphens between lowercase and uppercase letters (e.g., "camelCase" -> "camel-case").
	re2 := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	kebab = re2.ReplaceAllString(kebab, "${1}-${2}")

	// Convert to lowercase and trim leading/trailing hyphens.
	return strings.Trim(strings.ToLower(kebab), "-")
}

// UploadFile uploads a single file to the specified directory.
// Optionally converts images to WebP format and generates a unique filename.
// Returns the unique filename or an error if the upload fails.
func UploadFile(file io.Reader, fileName, uploadDirectory string, convertToWebP bool) (string, error) {
	// Ensure the upload directory exists with appropriate permissions.
	log.Info("📁 Ensuring upload directory exists: " + uploadDirectory)
	if err := os.MkdirAll(uploadDirectory, os.ModePerm); err != nil {
		// Log and return an error if directory creation fails.
		log.Error("❌ Unable to create upload directory: " + err.Error())
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Convert the file to WebP format if requested and supported.
	if convertToWebP {
		log.Info("🖼️ Converting image to WebP format: " + fileName)
		converted, err := CheckAndConvertFile(file, fileName)
		if err != nil {
			// Log and return an error if WebP conversion fails.
			log.Error("❌ Conversion to WebP failed: " + err.Error())
			return "", err
		}
		file = converted
	}

	// Determine the file extension, updating to .webp if converted.
	ext := filepath.Ext(fileName)
	if convertToWebP && (ext == ".jpg" || ext == ".jpeg" || ext == ".png") {
		ext = ".webp"
	}
	// Extract the base filename without extension and convert to kebab-case.
	base := strings.TrimSuffix(filepath.Base(fileName), ext)
	baseKebab := toKebabCase(base)

	// Generate a unique filename using kebab-case base and a UUID.
	uniqueFilename := fmt.Sprintf("%s-%s%s", baseKebab, uuid.New().String(), ext)
	// Construct the full destination path.
	destinationPath := filepath.Join(uploadDirectory, uniqueFilename)

	// Create the destination file.
	log.Info("📝 Creating file: " + destinationPath)
	destination, err := os.Create(destinationPath)
	if err != nil {
		// Log and return an error if file creation fails.
		log.Error("❌ Failed to create file: " + err.Error())
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	// Copy the file content to the destination.
	log.Info("📤 Copying file content to destination")
	if _, err := io.Copy(destination, file); err != nil {
		// Log and return an error if content copying fails.
		log.Error("❌ Failed to write file content: " + err.Error())
		return "", fmt.Errorf("failed to copy file content to destination: %w", err)
	}

	// Log successful file upload.
	log.Success("✅ File uploaded successfully: " + uniqueFilename)
	return uniqueFilename, nil
}

// DeleteFile removes a single file from the specified directory.
// Returns an error if the file does not exist or deletion fails.
func DeleteFile(filename, uploadDirectory string) error {
	// Construct the full file path.
	filePath := filepath.Join(uploadDirectory, filename)
	// Log the start of the file deletion process.
	log.Info("🗑️ Deleting file: " + filePath)

	// Attempt to remove the file.
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			// Log and return an error if the file does not exist.
			log.Error("⚠️ File not found: " + filename)
			return fmt.Errorf("file not found: %w", err)
		}
		// Log and return an error for other deletion failures.
		log.Error("❌ Failed to delete file: " + err.Error())
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Log successful file deletion.
	log.Success("✅ File deleted: " + filename)
	return nil
}

// UploadMultipleFiles uploads multiple files and rolls back if any fail.
// Returns a list of uploaded filenames or an error if any upload fails.
func UploadMultipleFiles(files []io.Reader, fileNames []string, uploadDirectory string, convertToWebP bool) ([]string, error) {
	// Validate that the number of files matches the number of filenames.
	if len(files) != len(fileNames) {
		errMsg := "❌ Number of files and filenames mismatch"
		log.Error(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// Log the start of the batch upload process.
	log.Info("📦 Starting batch file upload")
	// Initialize a slice to store uploaded filenames.
	uploadedFiles := make([]string, 0, len(files))

	// Upload each file individually.
	for i, file := range files {
		// Attempt to upload the file.
		uniqueFilename, err := UploadFile(file, fileNames[i], uploadDirectory, convertToWebP)
		if err != nil {
			// Log the failure and initiate rollback of previously uploaded files.
			log.Error("❌ Upload failed for file: " + fileNames[i] + " — initiating rollback")
			for _, filename := range uploadedFiles {
				// Attempt to delete each successfully uploaded file during rollback.
				if delErr := DeleteFile(filename, uploadDirectory); delErr != nil {
					// Log if a rollback deletion fails.
					log.Error("⚠️ Rollback deletion failed for: " + filename + " — " + delErr.Error())
				}
			}
			// Return an error indicating which file failed.
			return nil, fmt.Errorf("failed to upload file %s: %w", fileNames[i], err)
		}
		// Add the uploaded filename to the list.
		uploadedFiles = append(uploadedFiles, uniqueFilename)
	}

	// Log successful batch upload.
	log.Success("✅ All files uploaded successfully")
	return uploadedFiles, nil
}

// DeleteMultipleFiles deletes multiple files and returns an error if any deletions fail.
// Returns nil if all deletions succeed.
func DeleteMultipleFiles(filenames []string, uploadDirectory string) error {
	// Log the start of the batch deletion process.
	log.Info("🗑️ Deleting multiple files")
	// Track any files that fail to delete.
	var failedDeletes []string

	// Attempt to delete each file.
	for _, filename := range filenames {
		if err := DeleteFile(filename, uploadDirectory); err != nil {
			// Log and collect filenames that fail to delete.
			log.Error("❌ Failed to delete file: " + filename + " — " + err.Error())
			failedDeletes = append(failedDeletes, filename)
		}
	}

	// Return an error if any deletions failed.
	if len(failedDeletes) > 0 {
		errMsg := fmt.Sprintf("⚠️ Could not delete files: %v", failedDeletes)
		log.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	// Log successful batch deletion.
	log.Success("✅ All files deleted successfully")
	return nil
}