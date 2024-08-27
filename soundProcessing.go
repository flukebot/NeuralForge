package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
		return fmt.Errorf("error creating sounds directory: %v", err)
	}

	projectData, err := a.GetProjectData(projectName)
	if err != nil {
		return fmt.Errorf("error getting project data: %v", err)
	}

	for folder, files := range projectData.FileList {
		for _, file := range files {
			sourceFilePath := filepath.Join(projectData.SelectedDirectory, folder, file)
			targetFilePath := filepath.Join(soundsDir, strings.TrimSuffix(file, filepath.Ext(file))+".wav")

			if strings.ToLower(filepath.Ext(file)) == ".wav" {
				err := copyFile(sourceFilePath, targetFilePath)
				if err != nil {
					return fmt.Errorf("error copying WAV file: %v", err)
				}
			} else {
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
		return nil, err
	}
	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)
	spectrogramsDir := filepath.Join(projectDir, "spectrograms")

	err = os.MkdirAll(spectrogramsDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("error creating spectrograms directory: %v", err)
	}

	projectData, err := a.GetProjectData(projectName)
	if err != nil {
		return nil, fmt.Errorf("error getting project data: %v", err)
	}

	duplicates := []string{}

	for folder, files := range projectData.FileList {
		for _, file := range files {
			sourceFilePath := filepath.Join(projectData.SelectedDirectory, folder, file)
			chunkFilePath := filepath.Join(spectrogramsDir, strings.TrimSuffix(file, filepath.Ext(file))+"_chunk.wav")

			err := detectAndChunkAudio(sourceFilePath, chunkFilePath)
			if err != nil {
				return nil, fmt.Errorf("error chunking audio file: %v", err)
			}

			spectrogramData, err := generateSpectrogramData(chunkFilePath)
			if err != nil {
				return nil, fmt.Errorf("error generating spectrogram data: %v", err)
			}

			md5Hash, err := calculateMD5FromData(spectrogramData)
			if err != nil {
				return nil, fmt.Errorf("error calculating MD5 hash: %v", err)
			}

			spectrogramJSON := SpectrogramData{
				FileName:    file,
				MD5Hash:     md5Hash,
				ChunkPath:   chunkFilePath,
				Spectrogram: spectrogramData,
			}
			jsonFilePath := filepath.Join(spectrogramsDir, md5Hash+".json")

			if _, err := os.Stat(jsonFilePath); os.IsNotExist(err) {
				err = saveJSON(jsonFilePath, spectrogramJSON)
				if err != nil {
					return nil, fmt.Errorf("error saving spectrogram data: %v", err)
				}
			} else {
				duplicates = append(duplicates, file)
			}
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

	cmd := exec.Command("ffmpeg", "-i", src, "-af", "silencedetect=n=-30dB:d=0.5", "-f", "null", "-", "2>", tempFile.Name())
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error running FFmpeg silence detection: %v", err)
	}

	chunks, err := parseSilenceDetection(tempFile.Name())
	if err != nil {
		return fmt.Errorf("error parsing silence detection output: %v", err)
	}

	if len(chunks) == 0 {
		return copyFile(src, dst)
	}

	for i, chunk := range chunks {
		chunkFilePath := strings.TrimSuffix(dst, filepath.Ext(dst)) + fmt.Sprintf("_chunk%d.wav", i+1)
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

func generateSpectrogramData(src string) ([][]float64, error) {
	tempFile, err := os.CreateTemp("", "spectrogram-*.txt")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	cmd := exec.Command("ffmpeg", "-i", src, "-lavfi", "showspectrumpic=s=1024x1024:legend=disabled:scale=cbrt", "-f", "data", tempFile.Name())
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error generating spectrogram with FFmpeg: %v", err)
	}

	return readSpectrogramData(tempFile.Name())
}

func readSpectrogramData(filePath string) ([][]float64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening spectrogram data file: %v", err)
	}
	defer file.Close()

	var spectrogram [][]float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		values := strings.Fields(line)
		var row []float64
		for _, value := range values {
			floatValue, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing spectrogram value: %v", err)
			}
			row = append(row, floatValue)
		}
		spectrogram = append(spectrogram, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading spectrogram data: %v", err)
	}

	return spectrogram, nil
}

func calculateMD5FromData(data [][]float64) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

func saveJSON(filePath string, data SpectrogramData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, jsonData, os.ModePerm)
}
