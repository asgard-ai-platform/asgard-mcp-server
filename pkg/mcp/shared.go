package mcp

import (
	"fmt"
	"net/http"
	"os"
)

const (
	UploadedFilePathsFieldName = "_uploaded_file_paths"
	FormDataKeyJSON            = "json"
	FormDataKeyFile            = "file"
)

var UploadedFilePathsSchema = map[string]interface{}{
	"type":        "array",
	"description": "List of file paths to be uploaded",
	"items": map[string]interface{}{
		"type":        "string",
		"description": "Path to the uploaded file",
	},
}

func DetectMime(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath) //nolint
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Read the first 512 bytes
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Detect the content type (MIME)
	mimeType := http.DetectContentType(buffer)
	return mimeType, nil
}
