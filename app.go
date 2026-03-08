package main

import (
	"context"
	"fmt"
	"sync"

	"GoDown/internal/downloader"
	"GoDown/internal/models"
	"GoDown/internal/platform"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	config *models.Config
}

func NewApp() *App {
	// Load config once when the app starts
	cfg, _ := models.LoadConfig("config.json")
	return &App{config: cfg}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectFolder opens the native OS picker and returns the path to Svelte
func (a *App) SelectFolder() string {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Select Download Directory",
		DefaultDirectory: "./downloads", // You can set a starting path here
	})

	if err != nil {
		return "Error selecting folder"
	}

	// change the config path to the selection
	a.config.DownloadPath = selection
	models.SaveConfig("config.json", a.config)

	// Returns the full path (e.g., C:\Users\Oliver\Downloads)
	return selection
}

func (a *App) UpdateConfig(newCfg models.Config) string {
	err := models.SaveConfig("config.json", &newCfg)
	if err != nil {
		return "Error saving: " + err.Error()
	}
	a.config = &newCfg
	return "Settings Saved!"
}

// GetConfig sends the current config to Svelte when it opens the settings
func (a *App) GetConfig() *models.Config {
	return a.config
}

// DownloadPlaylist is what your Svelte button will call
func (a *App) DownloadPlaylist(url string) string {
	tracks, err := platform.FetchMetadata(url)
	if err != nil {
		return fmt.Sprintf("Error fetching metadata: %v", err)
	}

	tracksAmount := len(tracks)
	jobs := make(chan models.Track, tracksAmount)
	results := make(chan error, tracksAmount)
	var wg sync.WaitGroup

	// Start workers based on your config
	for w := 1; w <= a.config.Workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for track := range jobs {
				// We use context.Background() for now to keep it simple
				downloader.DownloadTrack(context.Background(), track, a.config)
				results <- nil // Simplified for this basic version
			}
		}()
	}

	// Feed tracks to workers
	for _, t := range tracks {
		jobs <- t
	}
	close(jobs)

	// Wait for completion in a goroutine so we don't block the UI thread
	go func() {
		wg.Wait()
		close(results)
	}()

	return fmt.Sprintf("Started downloading %d tracks...", tracksAmount)
}
