package main

import (
	"fmt"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const batchSize = 200

/*
type SpectrogramData struct {
	FileName    string      `json:"file_name"`
	MD5Hash     string      `json:"md5_hash"`
	Spectrogram [][]float64 `json:"spectrogram"`
}*/

func (a *App) ConvertFilesToWAV(projectName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)
	soundsDir := filepath.Join(projectDir, "sounds")

	err = os.MkdirAll(soundsDir, os.ModePerm)
	if err != nil {
		return a.LogError(projectName, err, "error creating sounds directory")
	}

	projectData, err := a.GetProjectData(projectName)
	if err != nil {
		return a.LogError(projectName, err, "error getting project data")
	}

	var mutex sync.Mutex
	errorList := []error{}

	files := []string{}
	for folder, fileList := range projectData.FileList {
		for _, file := range fileList {
			files = append(files, filepath.Join(folder, file))
		}
	}

	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		var wg sync.WaitGroup
		wg.Add(len(files[i:end]))

		for _, file := range files[i:end] {
			go func(file string) {
				defer wg.Done()
				sourceFilePath := filepath.Join(projectData.SelectedDirectory, file)
				// Use only the base file name for the target file path
				targetFileName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)) + ".wav"
				targetFilePath := filepath.Join(soundsDir, targetFileName)

				if _, err := os.Stat(targetFilePath); err == nil {
					a.LogError(projectName, nil, fmt.Sprintf("WAV file already exists: %s", targetFilePath))
					return
				}

				if strings.ToLower(filepath.Ext(file)) == ".wav" {
					err = copyFile(sourceFilePath, targetFilePath)
				} else {
					err = convertToWAV(sourceFilePath, targetFilePath)
				}

				if err != nil {
					mutex.Lock()
					errorList = append(errorList, a.LogError(projectName, err, fmt.Sprintf("error processing file: %s", sourceFilePath)))
					mutex.Unlock()
				}
			}(file)
		}
		wg.Wait()
	}

	if len(errorList) > 0 {
		return fmt.Errorf("errors occurred during file conversion")
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
	cmd := exec.Command("ffmpeg", "-i", src, dst)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error converting file to WAV: %v. Output: %s\n", err, string(output))
		return err
	}
	return nil
}

func (a *App) LogError(projectName string, err error, message string) error {
	homeDir, _ := os.UserHomeDir()
	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)
	logFilePath := filepath.Join(projectDir, "log.error")

	f, errFile := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errFile != nil {
		fmt.Printf("Failed to open log file: %v\n", errFile)
		return errFile
	}
	defer f.Close()

	logMessage := message
	if err != nil {
		logMessage += fmt.Sprintf(": %v", err)
	}
	logMessage += "\n"

	if _, errFile = f.WriteString(logMessage); errFile != nil {
		fmt.Printf("Failed to write to log file: %v\n", errFile)
	}

	return err
}
