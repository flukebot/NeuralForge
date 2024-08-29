package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type SpectrogramData struct {
	FileName    string      `json:"file_name"`
	MD5Hash     string      `json:"md5_hash"`
	ChunkPath   string      `json:"chunk_path"`
	Spectrogram [][]float64 `json:"spectrogram"`
}

const spectrogramChunkBatch = 10 // Number of files to process concurrently

func (a *App) ProcessAudioChunksAndSpectrograms(projectName string) ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return nil, err
	}
	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)
	soundsDir := filepath.Join(projectDir, "sounds")
	spectrogramsDir := filepath.Join(projectDir, "spectrograms")

	err = os.MkdirAll(spectrogramsDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating spectrograms directory:", err)
		return nil, a.LogError(projectName, err, "error creating spectrograms directory")
	}

	var duplicates []string
	var wg sync.WaitGroup
	var fileCounter int32
	fileChan := make(chan string, spectrogramChunkBatch)
	duplicateChan := make(chan string, spectrogramChunkBatch)

	// Worker to process files
	for i := 0; i < spectrogramChunkBatch; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				md5Hash, err := a.processWAVFile(filePath, spectrogramsDir)
				if err != nil {
					a.LogError(projectName, err, fmt.Sprintf("error processing WAV file: %s", filePath))
				} else {
					duplicateChan <- md5Hash
				}
				atomic.AddInt32(&fileCounter, 1)
				fmt.Printf("Processed file %d: %s\n", fileCounter, filePath)
			}
		}()
	}

	// Collect duplicates
	go func() {
		for md5Hash := range duplicateChan {
			duplicates = append(duplicates, md5Hash)
		}
	}()

	// Walk through all .wav files and send them to the workers
	err = filepath.Walk(soundsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".wav") {
			fileChan <- path
		}
		return nil
	})

	if err != nil {
		close(fileChan)
		close(duplicateChan)
		wg.Wait()
		return nil, err
	}

	// Close channels and wait for all workers to finish
	close(fileChan)
	wg.Wait()
	close(duplicateChan)

	fmt.Println("Audio processing completed with spectrogram generation.")
	return duplicates, nil
}

func (a *App) processWAVFile(filePath, spectrogramsDir string) (string, error) {
	chunkFilePath := filePath

	md5Hash, err := generateAndSaveSpectrogramData(chunkFilePath, spectrogramsDir)
	if err != nil {
		return "", fmt.Errorf("error processing spectrogram for chunk: %s", chunkFilePath)
	}

	// Delete the chunked .wav file after processing
	err = os.Remove(chunkFilePath)
	if err != nil {
		return "", fmt.Errorf("error deleting chunk file: %s", chunkFilePath)
	}

	return md5Hash, nil
}

func generateAndSaveSpectrogramData(chunkFilePath, spectrogramsDir string) (string, error) {
	md5Hash, err := calculateMD5FromFile(chunkFilePath)
	if err != nil {
		return "", fmt.Errorf("error calculating MD5 hash for chunk: %v", err)
	}

	jsonFilePath := filepath.Join(spectrogramsDir, md5Hash+".json")

	if _, err := os.Stat(jsonFilePath); err == nil {
		return md5Hash, nil
	}

	spectrogramData, err := generateSpectrogramData(chunkFilePath)
	if err != nil {
		return "", fmt.Errorf("error generating spectrogram data: %v", err)
	}

	spectrogramJSON := SpectrogramData{
		FileName:    filepath.Base(chunkFilePath),
		MD5Hash:     md5Hash,
		ChunkPath:   chunkFilePath,
		Spectrogram: spectrogramData,
	}

	err = saveJSON(jsonFilePath, spectrogramJSON)
	if err != nil {
		return "", fmt.Errorf("error saving spectrogram data: %v", err)
	}

	return md5Hash, nil
}

func generateSpectrogramData(src string) ([][]float64, error) {
	tempFile, err := os.CreateTemp("", "spectrogram-*.png")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	cmd := exec.Command("ffmpeg", "-y", "-i", src, "-lavfi", "showspectrumpic=s=1024x1024:legend=disabled:scale=cbrt", tempFile.Name())

	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("FFmpeg error output:\n%s\n", string(cmdOutput))
		return nil, fmt.Errorf("error generating spectrogram with FFmpeg: %v", err)
	}

	return extractSpectrogramDataFromImage(tempFile.Name())
}

func extractSpectrogramDataFromImage(imagePath string) ([][]float64, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error opening spectrogram image: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding spectrogram image: %v", err)
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	spectrogramData := make([][]float64, height)

	for y := 0; y < height; y++ {
		row := make([]float64, width)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			intensity := float64(r+g+b) / (3 * 65535.0)
			row[x] = intensity
		}
		spectrogramData[y] = row
	}

	return spectrogramData, nil
}

func calculateMD5FromFile(filePath string) (string, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(fileData)
	return hex.EncodeToString(hash[:]), nil
}

func saveJSON(filePath string, data SpectrogramData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, jsonData, os.ModePerm)
}


