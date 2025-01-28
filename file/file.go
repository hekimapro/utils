package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// UploadFile handles the uploading of a single file.
func UploadFile(file io.Reader, handlerFilename, uploadDirectory string) (string, error) {
	// Ensure the upload directory exists.
	err := os.MkdirAll(uploadDirectory, os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate a unique filename with the original file extension.
	uniqueFilename := uuid.New().String() + filepath.Ext(handlerFilename)
	destinationPath := filepath.Join(uploadDirectory, uniqueFilename)

	// Create the destination file.
	destination, err := os.Create(destinationPath)
	if err != nil {
		fmt.Println(err.Error())
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	// Copy file content to the destination.
	_, err = io.Copy(destination, file)
	if err != nil {
		fmt.Println(err.Error())
		return "", fmt.Errorf("failed to copy file content to destination: %w", err)
	}

	// Return the unique filename.
	return uniqueFilename, nil
}

// DeleteFile removes a single file from the specified directory.
func DeleteFile(filename, uploadDirectory string) error {
	filePath := filepath.Join(uploadDirectory, filename)

	err := os.Remove(filePath)
	if err != nil {
		fmt.Println(err.Error())
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %w", err)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// UploadMultipleFiles handles the uploading of multiple files.
// files: A slice of file readers and their original filenames.
// uploadDirectory: The directory where files should be uploaded.
// Returns a slice of unique filenames or an error if the upload fails for any file.
func UploadMultipleFiles(files []io.Reader, handlerFilenames []string, uploadDirectory string) ([]string, error) {
	if len(files) != len(handlerFilenames) {
		return nil, fmt.Errorf("mismatch between number of files and filenames")
	}

	uploadedFiles := make([]string, 0, len(files))

	for i, file := range files {
		uniqueFilename, err := UploadFile(file, handlerFilenames[i], uploadDirectory)
		if err != nil {
			// Rollback: Delete already uploaded files on failure.
			for _, filename := range uploadedFiles {
				_ = DeleteFile(filename, uploadDirectory)
			}
			return nil, fmt.Errorf("failed to upload file %s: %w", handlerFilenames[i], err)
		}
		uploadedFiles = append(uploadedFiles, uniqueFilename)
	}

	return uploadedFiles, nil
}

// DeleteMultipleFiles removes multiple files from the specified directory.
// filenames: A slice of filenames to be deleted.
// uploadDirectory: The directory where files are stored.
// Returns an error if any file fails to delete, but continues attempting deletion for others.
func DeleteMultipleFiles(filenames []string, uploadDirectory string) error {
	var failedDeletes []string

	for _, filename := range filenames {
		if err := DeleteFile(filename, uploadDirectory); err != nil {
			fmt.Printf("Failed to delete file %s: %s\n", filename, err.Error())
			failedDeletes = append(failedDeletes, filename)
		}
	}

	if len(failedDeletes) > 0 {
		return fmt.Errorf("failed to delete files: %v", failedDeletes)
	}

	return nil
}
