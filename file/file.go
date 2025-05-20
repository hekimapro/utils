package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hekimapro/utils/log"
)

// toKebabCase converts a string to kebab-case (lowercase with hyphens)
func toKebabCase(stringValue string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	kebab := re.ReplaceAllString(stringValue, "-")

	re2 := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	kebab = re2.ReplaceAllString(kebab, "${1}-${2}")

	return strings.Trim(strings.ToLower(kebab), "-")
}

// UploadFile uploads a single file to the specified directory
func UploadFile(file io.Reader, fileName, uploadDirectory string, convertToWebP bool) (string, error) {
	log.Info("Creating upload directory if not exists: " + uploadDirectory)
	err := os.MkdirAll(uploadDirectory, os.ModePerm)
	if err != nil {
		log.Error("Failed to create upload directory: " + err.Error())
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	if convertToWebP {
		log.Info("Converting file to WebP: " + fileName)
		file, err = CheckAndConvertFile(file, fileName)
		if err != nil {
			log.Error("Conversion to WebP failed: " + err.Error())
			return "", err
		}
	}

	ext := filepath.Ext(fileName)
	if convertToWebP {
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			ext = ".webp"
		}
	}
	base := strings.TrimSuffix(filepath.Base(fileName), ext)
	baseKebab := toKebabCase(base)

	uniqueFilename := fmt.Sprintf("%s-%s%s", baseKebab, uuid.New().String(), ext)
	destinationPath := filepath.Join(uploadDirectory, uniqueFilename)

	log.Info("Creating destination file: " + destinationPath)
	destination, err := os.Create(destinationPath)
	if err != nil {
		log.Error("Failed to create destination file: " + err.Error())
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	log.Info("Copying file content to destination")
	_, err = io.Copy(destination, file)
	if err != nil {
		log.Error("Failed to copy file content: " + err.Error())
		return "", fmt.Errorf("failed to copy file content to destination: %w", err)
	}

	log.Success("File uploaded successfully: " + uniqueFilename)
	return uniqueFilename, nil
}

// DeleteFile removes a single file from the specified directory
func DeleteFile(filename, uploadDirectory string) error {
	filePath := filepath.Join(uploadDirectory, filename)

	log.Info("Deleting file: " + filePath)
	err := os.Remove(filePath)
	if err != nil {
		log.Error("Failed to delete file: " + err.Error())
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %w", err)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	log.Success("File deleted successfully: " + filename)
	return nil
}

// UploadMultipleFiles uploads multiple files and rolls back if any fail
func UploadMultipleFiles(files []io.Reader, fileNames []string, uploadDirectory string, convertToWebP bool) ([]string, error) {
	if len(files) != len(fileNames) {
		errMsg := "Mismatch between number of files and filenames"
		log.Error(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	uploadedFiles := make([]string, 0, len(files))
	for i, file := range files {
		uniqueFilename, err := UploadFile(file, fileNames[i], uploadDirectory, convertToWebP)
		if err != nil {
			log.Error("Failed to upload file: " + fileNames[i] + " - rolling back")
			for _, filename := range uploadedFiles {
				if delErr := DeleteFile(filename, uploadDirectory); delErr != nil {
					log.Error("Rollback deletion failed for file: " + filename + " - " + delErr.Error())
				}
			}
			return nil, fmt.Errorf("failed to upload file %s: %w", fileNames[i], err)
		}
		uploadedFiles = append(uploadedFiles, uniqueFilename)
	}

	log.Success("All files uploaded successfully")
	return uploadedFiles, nil
}

// DeleteMultipleFiles deletes multiple files and returns error if any deletions fail
func DeleteMultipleFiles(filenames []string, uploadDirectory string) error {
	var failedDeletes []string

	for _, filename := range filenames {
		err := DeleteFile(filename, uploadDirectory)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to delete file %s: %s", filename, err.Error()))
			failedDeletes = append(failedDeletes, filename)
		}
	}

	if len(failedDeletes) > 0 {
		errMsg := fmt.Sprintf("Failed to delete files: %v", failedDeletes)
		log.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Success("All files deleted successfully")
	return nil
}
