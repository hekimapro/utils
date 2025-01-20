package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// File Upload
func UploadFile(file io.Reader, handlerFilename, uploadDirectory string) (string, error) {

	// check for upload directory
	err := os.MkdirAll(uploadDirectory, os.ModePerm)

	if err != nil {
		fmt.Println(err.Error())
		return "", CreateError("failed to create upload directory")
	}

	// generate a unique filename
	uniqueFilename := uuid.New().String() + filepath.Ext(handlerFilename)

	// create a destination file
	destinationPath := filepath.Join(uploadDirectory, uniqueFilename)
	destination, err := os.Create(destinationPath)

	if err != nil {
		fmt.Println(err.Error())
		return "", CreateError("failed to create destination file")
	}

	defer destination.Close()

	// copy the file content to the destination
	_, err = io.Copy(destination, file)

	if err != nil {
		fmt.Println(err.Error())
		return "", CreateError("failed to copy file content to the destination")
	}

	return uniqueFilename, nil
}

// Delete Uploaded File
func DeleteFile(filename, uploadDirectory string) error {

	// construct the full file path
	filePath := filepath.Join(uploadDirectory, filename)

	// remove the file
	err := os.Remove(filePath)

	if err != nil {
		fmt.Println(err.Error())
		if os.IsNotExist(err) {
			return CreateError("file not found")
		}
		return CreateError("failed to delete file")
	}

	return nil
}
