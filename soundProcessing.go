// sound_processing.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (a *App) ConvertFilesToWAV(projectName string) error {
	// Get project directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)
	soundsDir := filepath.Join(projectDir, "sounds")

	// Ensure the "sounds" subdirectory exists
	err = os.MkdirAll(soundsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating sounds directory: %v", err)
	}

	// Get the list of files from the project data
	projectData, err := a.GetProjectData(projectName)
	if err != nil {
		return fmt.Errorf("error getting project data: %v", err)
	}

	// Loop through all files and convert or copy them to the sounds directory
	for folder, files := range projectData.FileList {
		for _, file := range files {
			sourceFilePath := filepath.Join(projectData.SelectedDirectory, folder, file)
			targetFilePath := filepath.Join(soundsDir, strings.TrimSuffix(file, filepath.Ext(file))+".wav")

			// Check if the file is already a WAV file
			if strings.ToLower(filepath.Ext(file)) == ".wav" {
				err := copyFile(sourceFilePath, targetFilePath)
				if err != nil {
					return fmt.Errorf("error copying WAV file: %v", err)
				}
			} else {
				// Convert to WAV using FFmpeg
				err := convertToWAV(sourceFilePath, targetFilePath)
				if err != nil {
					return fmt.Errorf("error converting file to WAV: %v", err)
				}
			}
		}
	}

	fmt.Println("All files have been processed and converted to WAV format.")
	return nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func convertToWAV(src, dst string) error {
	// Use FFmpeg to convert the file to WAV
	cmd := exec.Command("ffmpeg", "-i", src, dst)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error converting file to WAV: %v", err)
	}
	return nil
}
