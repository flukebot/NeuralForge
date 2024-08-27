package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type App struct {
    ctx context.Context
}

type ProjectData struct {
	SelectedDirectory string            `json:"selected_directory"`
	FileList          map[string][]string `json:"file_list"`
}

func NewApp() *App {
    return &App{}
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx

    // Initialize project directories
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

func (a *App) Greet(name string) string {
    return fmt.Sprintf("Hello %s, It's show time!", name)
}

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


func (a *App) SaveSelectedDirectory(directoryPath string, projectName string) error {
	// Save the selected directory path in the project folder config JSON
	projectDir, err := a.CreateProject(projectName)
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(projectDir, "config.json")
	configData := map[string]string{
		"selected_directory": directoryPath,
	}

	configJson, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configFilePath, configJson, os.ModePerm)
	if err != nil {
		return err
	}

	// Get the list of files and subfiles in JSON format
	fileListJson, err := a.ListFilesInDirectory(directoryPath)
	if err != nil {
		return err
	}

	// Save the file list JSON in a separate file in the project folder
	fileListFilePath := filepath.Join(projectDir, "file_list.json")
	err = ioutil.WriteFile(fileListFilePath, []byte(fileListJson), os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Println("Selected directory and file list have been saved successfully.")
	return nil
}


func (a *App) GetProjectData(projectName string) (*ProjectData, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projectDir := filepath.Join(homeDir, "NeuralForge", "projects", projectName)

	// Read config.json to get the selected directory path
	configFilePath := filepath.Join(projectDir, "config.json")
	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config map[string]string
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config data: %v", err)
	}

	selectedDirectory, ok := config["selected_directory"]
	if !ok {
		return nil, fmt.Errorf("selected_directory not found in config")
	}

	// Read file_list.json to get the list of files
	fileListFilePath := filepath.Join(projectDir, "file_list.json")
	fileListData, err := ioutil.ReadFile(fileListFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file list: %v", err)
	}

	var fileList map[string][]string
	err = json.Unmarshal(fileListData, &fileList)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling file list: %v", err)
	}

	// Combine the data into a ProjectData struct
	projectData := &ProjectData{
		SelectedDirectory: selectedDirectory,
		FileList:          fileList,
	}

	return projectData, nil
}

