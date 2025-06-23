package tasks

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
)

func DownloadFile(url string) (TaskResult, error) {
	fileName := filepath.Base(url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	result := DownloadResult{
		Name:     fileName,
		Bytes:    data,
		FileType: filepath.Ext(fileName),
	}

	return &result, nil
}
