package main

import (
	"context"
	"fmt"
	"strings"
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

func (a *App) DownloadPlaylist(url string) string {
	tracks, err := platform.FetchMetadata(url)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Use a helper to generate consistent IDs
	getID := func(t models.Track) string {
		return fmt.Sprintf("%s-%s-%d", t.Title, strings.Join(t.Artists, ","), t.DurationSec)
	}

	// 1. Queue them all immediately so the UI creates the divs
	for _, t := range tracks {
		runtime.EventsEmit(a.ctx, "download_progress", map[string]interface{}{
			"id":     getID(t),
			"title":  t.Title,
			"status": "Queued",
		})
	}

	// 2. Wrap the worker logic in a goroutine so the function returns immediately
	go func() {
		jobs := make(chan models.Track, len(tracks))
		var wg sync.WaitGroup

		for w := 1; w <= a.config.Workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for track := range jobs {
					trackID := getID(track) // MUST match the ID above

					// Update to Downloading
					runtime.EventsEmit(a.ctx, "download_progress", map[string]interface{}{
						"id":     trackID,
						"status": "Downloading...",
					})

					err := downloader.DownloadTrack(context.Background(), track, a.config)

					finalStatus := "Completed"
					if err != nil {
						finalStatus = "Failed"
					}

					// Update to Finished
					runtime.EventsEmit(a.ctx, "download_progress", map[string]interface{}{
						"id":     trackID,
						"status": finalStatus,
					})
				}
			}()
		}

		for _, t := range tracks {
			jobs <- t
		}
		close(jobs)
		wg.Wait()
	}()

	return fmt.Sprintf("Queued %d tracks", len(tracks))
}
