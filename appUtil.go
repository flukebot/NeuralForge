package main

import "github.com/wailsapp/wails/v2/pkg/runtime"



func (a *App) OpenDirectoryDialog() (string, error) {
	// This function will open a directory picker dialog and return the selected directory
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
}