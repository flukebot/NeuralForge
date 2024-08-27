package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)



func (a *App) OpenDirectoryDialog() (string, error) {
	// This function will open a directory picker dialog and return the selected directory
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
}

func (a *App) ListFilesInDirectory(dirPath string) (string, error) {
	fileList := make(map[string][]string)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relativePath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			dir := filepath.Dir(relativePath)
			fileList[dir] = append(fileList[dir], filepath.Base(relativePath))
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// Convert the fileList map to JSON
	jsonData, err := json.MarshalIndent(fileList, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}