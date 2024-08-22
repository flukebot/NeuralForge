package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Ensure the NeuralForge folder structure exists
	projectDir, err := a.CreateProjectFolder()
	if err != nil {
		fmt.Println("Error creating NeuralForge directory:", err)
		return
	}

	// Ensure the "projects" subdirectory exists
	projectsDir := filepath.Join(projectDir, "projects")
	err = os.MkdirAll(projectsDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating projects directory:", err)
		return
	}

	fmt.Println("NeuralForge and projects directories are ready.")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

//-------------------------------

// CreateNeuralForgeFolder creates a folder called "NeuralForge" in the user's home directory
func (a *App) CreateProjectFolder() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	NeuralForgeDir := filepath.Join(homeDir, "NeuralForge")

	err = os.MkdirAll(NeuralForgeDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	return NeuralForgeDir, nil
}

// ListProjects lists all folders in the NeuralForge/projects/ directory
func (a *App) ListProjects() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(homeDir, "NeuralForge", "projects")
	folders, err := ioutil.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var projectNames []string
	for _, folder := range folders {
		if folder.IsDir() {
			projectNames = append(projectNames, folder.Name())
		}
	}

	return projectNames, nil
}

// CreateProject creates a new project folder in the user's NeuralForge/projects/ directory
func (a *App) CreateProject(projectName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)
	err = os.MkdirAll(projectDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	return projectDir, nil
}
