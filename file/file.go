package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// toKebabCase converts a string to kebab-case (lowercase with hyphens)
// Replaces non-alphanumeric characters with hyphens, adds hyphens between case transitions, and trims edges
func toKebabCase(stringValue string) string {
	// Replace all non-alphanumeric characters with hyphens
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	kebab := re.ReplaceAllString(stringValue, "-")

	// Add hyphen between lowercase-to-uppercase transitions
	re2 := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	kebab = re2.ReplaceAllString(kebab, "${1}-${2}")

	// Convert to lowercase and trim hyphens
	return strings.Trim(strings.ToLower(kebab), "-")
}

// UploadFile uploads a single file to the specified directory
// Optionally converts images to WebP, generates a unique kebab-case filename with a UUID, and copies the file content
// Returns the unique filename or an error if the upload fails
func UploadFile(file io.Reader, fileName, uploadDirectory string, convertToWebP bool) (string, error) {
	// Ensure the upload directory exists with proper permissions
	err := os.MkdirAll(uploadDirectory, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Convert the file to WebP if requested and supported
	if convertToWebP {
		file, err = CheckAndConvertFile(file, fileName)
		if err != nil {
			return "", err
		}
	}

	// Extract file extension and base name from the provided filename
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(filepath.Base(fileName), ext)

	// Convert base name to kebab-case for consistency
	baseKebab := toKebabCase(base)

	// Generate a unique filename using kebab-case base, UUID, and extension
	uniqueFilename := fmt.Sprintf("%s-%s%s", baseKebab, uuid.New().String(), ext)
	destinationPath := filepath.Join(uploadDirectory, uniqueFilename)

	// Create the destination file in the upload directory
	destination, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	// Copy the file content from the reader to the destination file
	_, err = io.Copy(destination, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content to destination: %w", err)
	}

	// Return the unique filename for reference
	return uniqueFilename, nil
}

// DeleteFile removes a single file from the specified directory
// Takes the filename and directory path, deletes the file, and handles errors
// Returns an error if the file does not exist or deletion fails
func DeleteFile(filename, uploadDirectory string) error {
	// Construct the full file path
	filePath := filepath.Join(uploadDirectory, filename)

	// Attempt to remove the file
	err := os.Remove(filePath)
	if err != nil {
		// Log the error to stdout
		fmt.Println(err.Error())
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %w", err)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Return nil on successful deletion
	return nil
}

// UploadMultipleFiles handles the uploading of multiple files to the specified directory
// Optionally converts images to WebP, uploads each file, and ensures consistency
// Returns a slice of unique filenames or an error, rolling back on failure
func UploadMultipleFiles(files []io.Reader, fileNames []string, uploadDirectory string, convertToWebP bool) ([]string, error) {
	// Validate that the number of files matches the number of filenames
	if len(files) != len(fileNames) {
		return nil, fmt.Errorf("mismatch between number of files and filenames")
	}

	// Initialize slice to store uploaded filenames
	uploadedFiles := make([]string, 0, len(files))

	// Iterate through files and upload each one
	for i, file := range files {
		uniqueFilename, err := UploadFile(file, fileNames[i], uploadDirectory, convertToWebP)
		if err != nil {
			// Rollback: Delete already uploaded files to maintain consistency
			for _, filename := range uploadedFiles {
				_ = DeleteFile(filename, uploadDirectory)
			}
			return nil, fmt.Errorf("failed to upload file %s: %w", fileNames[i], err)
		}
		uploadedFiles = append(uploadedFiles, uniqueFilename)
	}

	// Return the list of successfully uploaded filenames
	return uploadedFiles, nil
}

// DeleteMultipleFiles removes multiple files from the specified directory
// Takes a slice of filenames and the directory path, attempts to delete each file
// Continues deletion even if some fail, returns an error with failed filenames
func DeleteMultipleFiles(filenames []string, uploadDirectory string) error {
	// Initialize slice to track failed deletions
	var failedDeletes []string

	// Iterate through filenames and attempt to delete each file
	for _, filename := range filenames {
		if err := DeleteFile(filename, uploadDirectory); err != nil {
			// Log the failure and track the failed filename
			fmt.Printf("Failed to delete file %s: %s\n", filename, err.Error())
			failedDeletes = append(failedDeletes, filename)
		}
	}

	// Return an error if any deletions failed, listing the failed filenames
	if len(failedDeletes) > 0 {
		return fmt.Errorf("failed to delete files: %v", failedDeletes)
	}

	// Return nil on successful deletion
	return nil
}
