package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// UploadFile handles the uploading of a file.
// file: The file content as an io.Reader.
// handlerFilename: The original filename of the uploaded file.
// uploadDirectory: The directory where the file should be uploaded.
// Returns the unique filename generated for the uploaded file, or an error if the upload fails.
func UploadFile(file io.Reader, handlerFilename, uploadDirectory string) (string, error) {

	// Ensure the upload directory exists by creating it if necessary.
	err := os.MkdirAll(uploadDirectory, os.ModePerm)
	if err != nil {
		// Log and return an error if the directory cannot be created.
		fmt.Println(err.Error())
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate a unique filename by combining a UUID with the original file extension.
	uniqueFilename := uuid.New().String() + filepath.Ext(handlerFilename)

	// Define the destination file path.
	destinationPath := filepath.Join(uploadDirectory, uniqueFilename)

	// Create the destination file.
	destination, err := os.Create(destinationPath)
	if err != nil {
		// Log and return an error if the destination file cannot be created.
		fmt.Println(err.Error())
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	// Copy the content of the uploaded file to the destination file.
	_, err = io.Copy(destination, file)
	if err != nil {
		// Log and return an error if file content copying fails.
		fmt.Println(err.Error())
		return "", fmt.Errorf("failed to copy file content to the destination: %w", err)
	}

	// Return the unique filename assigned to the uploaded file.
	return uniqueFilename, nil
}

// DeleteFile removes an uploaded file from the specified directory.
// filename: The unique filename of the file to be deleted.
// uploadDirectory: The directory where the file is stored.
// Returns an error if the file cannot be deleted.
func DeleteFile(filename, uploadDirectory string) error {

	// Construct the full path of the file to be deleted.
	filePath := filepath.Join(uploadDirectory, filename)

	// Attempt to delete the file.
	err := os.Remove(filePath)
	if err != nil {
		// Log and return an error if the file cannot be found or deleted.
		fmt.Println(err.Error())
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %w", err)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Return nil if the file was successfully deleted.
	return nil
}
