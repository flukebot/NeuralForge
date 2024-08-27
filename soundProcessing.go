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
		fmt.Println("Error getting user home directory:", err)
		return nil, err
	}
	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)
	spectrogramsDir := filepath.Join(projectDir, "spectrograms")

	err = os.MkdirAll(spectrogramsDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating spectrograms directory:", err)
		return nil, fmt.Errorf("error creating spectrograms directory: %v", err)
	}

	projectData, err := a.GetProjectData(projectName)
	if err != nil {
		fmt.Println("Error getting project data:", err)
		return nil, fmt.Errorf("error getting project data: %v", err)
	}

	duplicates := []string{}

	for folder, files := range projectData.FileList {
		for _, file := range files {
			sourceFilePath := filepath.Join(projectData.SelectedDirectory, folder, file)
			chunkFilePath := filepath.Join(spectrogramsDir, strings.TrimSuffix(file, filepath.Ext(file))+"_chunk.wav")

			// Detect chunks and create spectrogram data
			err := detectAndChunkAudio(sourceFilePath, chunkFilePath)
			if err != nil {
				fmt.Println("Error chunking audio file:", err)
				return nil, fmt.Errorf("error chunking audio file: %v", err)
			}

			// Generate spectrogram data
			spectrogramData, err := generateSpectrogramData(chunkFilePath)
			if err != nil {
				fmt.Println("Error generating spectrogram data:", err)
				return nil, fmt.Errorf("error generating spectrogram data: %v", err)
			}

			// Calculate MD5 hash of the spectrogram data
			md5Hash, err := calculateMD5FromData(spectrogramData)
			if err != nil {
				fmt.Println("Error calculating MD5 hash:", err)
				return nil, fmt.Errorf("error calculating MD5 hash: %v", err)
			}

			// Save spectrogram data in a JSON file
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
					fmt.Println("Error saving spectrogram data:", err)
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
	// Temporary file to hold the output of the silence detection
	tempFile, err := os.CreateTemp("", "ffmpeg-silencedetect-*.txt")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Run FFmpeg to detect silent sections and capture standard error
	cmd := exec.Command("ffmpeg", "-i", src, "-af", "silencedetect=n=-30dB:d=0.5", "-f", "null", "-")

	// Capture the output and error
	cmd.Stderr = tempFile

	// Execute the command
	err = cmd.Run()
	if err != nil {
		tempFile.Seek(0, 0)
		errorOutput, _ := ioutil.ReadAll(tempFile)
		fmt.Println("FFmpeg error output:", string(errorOutput))
		return fmt.Errorf("error running FFmpeg silence detection: %v", err)
	}

	// Parse the silence detection output to determine the start and end times for chunks
	chunks, err := parseSilenceDetection(tempFile.Name())
	if err != nil {
		return fmt.Errorf("error parsing silence detection output: %v", err)
	}

	// If no clear chunks are found, treat the entire file as one chunk
	if len(chunks) == 0 {
		return copyFile(src, dst)
	}

	// Otherwise, extract the detected chunks
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
    // Generate a temporary file to hold the spectrogram data
    tempFile, err := os.CreateTemp("", "spectrogram-*.png")
    if err != nil {
        return nil, fmt.Errorf("error creating temp file: %v", err)
    }
    defer os.Remove(tempFile.Name())

    // Run FFmpeg to generate the spectrogram data as an image
    cmd := exec.Command("ffmpeg", "-y", "-i", src, "-lavfi", "showspectrumpic=s=1024x1024:legend=disabled:scale=cbrt", tempFile.Name())
    
    // Capture the output and error
    cmdOutput, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("FFmpeg error output:\n%s\n", string(cmdOutput))
        return nil, fmt.Errorf("error generating spectrogram with FFmpeg: %v", err)
    }

    // Read and parse the spectrogram data from the image file
    return extractSpectrogramDataFromImage(tempFile.Name())
}


func extractSpectrogramDataFromImage(imagePath string) ([][]float64, error) {
    // Open the image file
    file, err := os.Open(imagePath)
    if err != nil {
        return nil, fmt.Errorf("error opening spectrogram image: %v", err)
    }
    defer file.Close()

    // Decode the image
    img, _, err := image.Decode(file)
    if err != nil {
        return nil, fmt.Errorf("error decoding spectrogram image: %v", err)
    }

    // Convert the image data to spectrogram data
    bounds := img.Bounds()
    width, height := bounds.Max.X, bounds.Max.Y
    spectrogramData := make([][]float64, height)

    for y := 0; y < height; y++ {
        row := make([]float64, width)
        for x := 0; x < width; x++ {
            r, g, b, _ := img.At(x, y).RGBA()
            // Convert the RGB values to a single float64 value representing intensity
            intensity := float64(r+g+b) / (3 * 65535.0)
            row[x] = intensity
        }
        spectrogramData[y] = row
    }

    return spectrogramData, nil
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
