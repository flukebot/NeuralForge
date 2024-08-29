package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gonum.org/v1/gonum/mat"
)

const (
	maxK          = 10 // Maximum number of clusters to try
	elbowBatchSize = 100 // Number of spectrograms to process concurrently
)

type ElbowResult struct {
	WCSSValues []float64 `json:"wcss_values"`
	OptimalK   int        `json:"optimal_k"`
}

func (a *App) CalculateOptimalClusters(projectName string) (int, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}
	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)
	spectrogramsDir := filepath.Join(projectDir, "spectrograms")

	files, err := filepath.Glob(filepath.Join(spectrogramsDir, "*.json"))
	if err != nil {
		return 0, fmt.Errorf("error listing spectrogram JSON files: %v", err)
	}

	if len(files) == 0 {
		return 0, fmt.Errorf("no spectrogram files found in %s", spectrogramsDir)
	}

	// Load spectrogram data
	spectrograms := [][]float64{}
	var wg sync.WaitGroup
	var mu sync.Mutex

	fileChan := make(chan string, elbowBatchSize)
	resultChan := make(chan []float64, elbowBatchSize)

	// Worker to load spectrogram data
	for i := 0; i < elbowBatchSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				data, err := loadSpectrogramData(filePath)
				if err != nil {
					fmt.Printf("error loading spectrogram data: %v\n", err)
					continue
				}
				resultChan <- data
			}
		}()
	}

	// Send files to the worker
	go func() {
		for _, file := range files {
			fileChan <- file
		}
		close(fileChan)
	}()

	// Collect results
	go func() {
		for result := range resultChan {
			mu.Lock()
			spectrograms = append(spectrograms, result)
			mu.Unlock()
		}
	}()

	wg.Wait()
	close(resultChan)

	if len(spectrograms) == 0 {
		return 0, fmt.Errorf("no valid spectrogram data found")
	}

	// Convert spectrograms to a matrix
	dataMatrix := mat.NewDense(len(spectrograms), len(spectrograms[0]), nil)
	for i, s := range spectrograms {
		dataMatrix.SetRow(i, s)
	}

	// Use the elbow method to determine the optimal number of clusters
	wcss, optimalK, err := elbowMethod(dataMatrix)
	if err != nil {
		return 0, fmt.Errorf("error calculating optimal number of clusters: %v", err)
	}

	fmt.Printf("Optimal number of clusters determined by elbow method: %d\n", optimalK)

	// Save the elbow results
	elbowResult := ElbowResult{
		WCSSValues: wcss,
		OptimalK:   optimalK,
	}

	err = saveElbowResults(projectDir, elbowResult)
	if err != nil {
		return 0, fmt.Errorf("error saving elbow results: %v", err)
	}

	return optimalK, nil
}

func saveElbowResults(projectDir string, result ElbowResult) error {
	filePath := filepath.Join(projectDir, "elbow_results.json")
	fileData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, fileData, os.ModePerm)
}

func loadSpectrogramData(filePath string) ([]float64, error) {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var spectrogram SpectrogramData
	err = json.Unmarshal(fileData, &spectrogram)
	if err != nil {
		return nil, err
	}

	flattened := []float64{}
	for _, row := range spectrogram.Spectrogram {
		flattened = append(flattened, row...)
	}

	return flattened, nil
}

func elbowMethod(data *mat.Dense) ([]float64, int, error) {
	wcss := make([]float64, maxK)
	rows, _ := data.Dims()

	for k := 1; k <= maxK; k++ {
		centroids, err := initializeCentroids(data, k)
		if err != nil {
			return nil, 0, err
		}

		clusters := make([]int, rows)
		for i := 0; i < 100; i++ { // Run k-means for a fixed number of iterations
			for j := 0; j < rows; j++ {
				clusters[j] = closestCentroid(data.RowView(j), centroids)
			}
			centroids = updateCentroids(data, clusters, k)
		}

		// Calculate the Within-Cluster-Sum of Squares (WCSS)
		for j := 0; j < rows; j++ {
			centroid := centroids.RowView(clusters[j])
			wcss[k-1] += distance(data.RowView(j), centroid)
		}
	}

	// Find the elbow point
	optimalK := 1
	for i := 1; i < maxK-1; i++ {
		angle := math.Abs(wcss[i+1]-wcss[i]) - math.Abs(wcss[i]-wcss[i-1])
		if angle > 0 {
			optimalK = i + 1
		}
	}

	return wcss, optimalK, nil
}

func initializeCentroids(data *mat.Dense, k int) (*mat.Dense, error) {
	rows, cols := data.Dims()
	centroids := mat.NewDense(k, cols, nil)

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < k; i++ {
		randIndex := rand.Intn(rows)
		centroids.SetRow(i, data.RawRowView(randIndex))
	}

	return centroids, nil
}

func closestCentroid(point mat.Vector, centroids *mat.Dense) int {
	closest := 0
	minDist := distance(point, centroids.RowView(0))

	for i := 1; i < centroids.RawMatrix().Rows; i++ {
		d := distance(point, centroids.RowView(i))
		if d < minDist {
			closest = i
			minDist = d
		}
	}

	return closest
}

func updateCentroids(data *mat.Dense, clusters []int, k int) *mat.Dense {
	_, cols := data.Dims()
	newCentroids := mat.NewDense(k, cols, nil)
	clusterSizes := make([]int, k)

	for i, cluster := range clusters {
		row := data.RowView(i)
		for j := 0; j < cols; j++ {
			newCentroids.Set(cluster, j, newCentroids.At(cluster, j)+row.AtVec(j))
		}
		clusterSizes[cluster]++
	}

	for i := 0; i < k; i++ {
		for j := 0; j < cols; j++ {
			newCentroids.Set(i, j, newCentroids.At(i, j)/float64(clusterSizes[i]))
		}
	}

	return newCentroids
}

func distance(a, b mat.Vector) float64 {
	var dist float64
	for i := 0; i < a.Len(); i++ {
		diff := a.AtVec(i) - b.AtVec(i)
		dist += diff * diff
	}
	return math.Sqrt(dist)
}
