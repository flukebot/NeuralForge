package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type SpectrogramData struct {
	FileName    string        `json:"file_name"`
	MD5Hash     string        `json:"md5_hash"`
	ChunkPath   string        `json:"chunk_path"`
	Spectrogram [][]float64   `json:"spectrogram"`
}

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

	for folder, files := range projectData.FileList {
		for _, file := range files {
			sourceFilePath := filepath.Join(projectData.SelectedDirectory, folder, file)
			targetFilePath := filepath.Join(soundsDir, strings.TrimSuffix(file, filepath.Ext(file))+".wav")

			if _, err := os.Stat(targetFilePath); err == nil {
				a.LogError(projectName, nil, fmt.Sprintf("WAV file already exists: %s", targetFilePath))
				continue
			}

			if strings.ToLower(filepath.Ext(file)) == ".wav" {
				err := copyFile(sourceFilePath, targetFilePath)
				if err != nil {
					a.LogError(projectName, err, fmt.Sprintf("error copying WAV file: %s", sourceFilePath))
					continue
				}
			} else {
				err := convertToWAV(sourceFilePath, targetFilePath)
				if err != nil {
					a.LogError(projectName, err, fmt.Sprintf("error converting file to WAV: %s", sourceFilePath))
					continue
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
	cmd := exec.Command("ffmpeg", "-i", src, dst)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error converting file to WAV: %v", err)
	}
	return nil
}

func (a *App) ProcessAudioChunksAndSpectrograms(projectName string) ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return nil, err
	}
	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)
	spectrogramsDir := filepath.Join(projectDir, "spectrograms")

	err = os.MkdirAll(spectrogramsDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating spectrograms directory:", err)
		return nil, a.LogError(projectName, err, "error creating spectrograms directory")
	}

	projectData, err := a.GetProjectData(projectName)
	if err != nil {
		fmt.Println("Error getting project data:", err)
		return nil, a.LogError(projectName, err, "error getting project data")
	}

	duplicates := []string{}

	for folder, files := range projectData.FileList {
		for _, file := range files {
			sourceFilePath := filepath.Join(projectData.SelectedDirectory, folder, file)
			chunkFilePath := filepath.Join(spectrogramsDir, strings.TrimSuffix(file, filepath.Ext(file))+"_chunk.wav")

			if _, err := os.Stat(chunkFilePath); err == nil {
				a.LogError(projectName, nil, fmt.Sprintf("Chunk file already exists: %s", chunkFilePath))
				continue
			}

			err := detectAndChunkAudio(sourceFilePath, chunkFilePath)
			if err != nil {
				a.LogError(projectName, err, fmt.Sprintf("error chunking audio file: %s", sourceFilePath))
				continue
			}

			md5Hash, err := generateAndSaveSpectrogramData(chunkFilePath, spectrogramsDir)
			if err != nil {
				a.LogError(projectName, err, fmt.Sprintf("error processing spectrogram for chunk: %s", chunkFilePath))
				continue
			}

			duplicates = append(duplicates, md5Hash)
		}
	}

	fmt.Println("Audio processing completed with spectrogram generation.")
	return duplicates, nil
}

func detectAndChunkAudio(src, dst string) error {
	tempFile, err := os.CreateTemp("", "ffmpeg-silencedetect-*.txt")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	cmd := exec.Command("ffmpeg", "-i", src, "-af", "silencedetect=n=-30dB:d=0.5", "-f", "null", "-")
	cmd.Stderr = tempFile

	err = cmd.Run()
	if err != nil {
		tempFile.Seek(0, 0)
		errorOutput, _ := ioutil.ReadAll(tempFile)
		fmt.Println("FFmpeg error output:", string(errorOutput))
		return fmt.Errorf("error running FFmpeg silence detection: %v", err)
	}

	chunks, err := parseSilenceDetection(tempFile.Name())
	if err != nil {
		return fmt.Errorf("error parsing silence detection output: %v", err)
	}

	if len(chunks) == 0 {
		if _, err := os.Stat(dst); err == nil {
			return nil
		}
		return copyFile(src, dst)
	}

	for i, chunk := range chunks {
		chunkFilePath := strings.TrimSuffix(dst, filepath.Ext(dst)) + fmt.Sprintf("_chunk%d.wav", i+1)
		if _, err := os.Stat(chunkFilePath); err == nil {
			continue
		}
		err := extractChunk(src, chunkFilePath, chunk.start, chunk.end)
		if err != nil {
			return fmt.Errorf("error extracting chunk %d: %v", i+1, err)
		}
	}

	return nil
}

type audioChunk struct {
	start float64
	end   float64
}

func parseSilenceDetection(filePath string) ([]audioChunk, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening silence detection file: %v", err)
	}
	defer file.Close()

	var chunks []audioChunk
	var currentChunk *audioChunk

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "silence_start") {
			if currentChunk != nil {
				chunks = append(chunks, *currentChunk)
				currentChunk = nil
			}
		} else if strings.Contains(line, "silence_end") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				endTime, err := strconv.ParseFloat(parts[3], 64)
				if err == nil {
					currentChunk = &audioChunk{start: endTime}
				}
			}
		}
	}

	if currentChunk != nil {
		chunks = append(chunks, *currentChunk)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading silence detection file: %v", err)
	}

	return chunks, nil
}

func extractChunk(src, dst string, start, end float64) error {
	duration := end - start

	cmd := exec.Command("ffmpeg", "-i", src, "-ss", fmt.Sprintf("%f", start), "-t", fmt.Sprintf("%f", duration), "-c", "copy", dst)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error extracting chunk: %v", err)
	}

	return nil
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
