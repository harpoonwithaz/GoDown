// Oliver T.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"spotify-playlist-downloader/internal/downloader"
	"spotify-playlist-downloader/internal/models"
	"spotify-playlist-downloader/internal/platform"
)

func main() {
	startNow := time.Now()

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Error getting URL: Please provide a URL after the execution call")
		return
	}

	URL := args[1]

	cfg, err := models.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}
	if (*cfg).Workers > 5 {
		fmt.Println("Error: Cannot have more than 5 workers due to rate limiting")
		return
	}

	playlistTracks, err := platform.FetchMetadata(URL)
	if err != nil {
		log.Fatalf("Failure getting playlist items: %v", err)
	}
	fmt.Println("Success getting playlist items")

	skipped := 0
	tracksAmount := len(playlistTracks)

	// create the channels
	jobs := make(chan models.Track, tracksAmount)
	results := make(chan error, tracksAmount)

	// start 3 workers
	for w := 1; w <= (*cfg).Workers; w++ {
		go func(workerID int) {
			for track := range jobs {
				fmt.Printf("Worker %d starting: %s\n", workerID, track.Title)
				err := downloader.DownloadTrack(context.Background(), track, cfg)
				if err != nil {
					skipped++
				}
				results <- err
			}
		}(w)
	}

	// send jobs to workers
	for _, track := range playlistTracks {
		jobs <- track
	}
	close(jobs)

	for i := 0; i < len(playlistTracks); i++ {
		err := <-results
		if err != nil {
			fmt.Println("A download failed:", err)
		}
	}

	fmt.Printf("Finished downloading. Downloaded %v/%v tracks.\n", tracksAmount-skipped, tracksAmount)
	fmt.Printf("Time to download: %v", time.Since(startNow))
}
