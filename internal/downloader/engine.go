package downloader

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"GoDown/internal/models"
)

func DownloadTrack(ctx context.Context, track models.Track, cfg *models.Config) error {
	maxRetries := len((*cfg).Retries)

	for i, retry := range (*cfg).Retries {
		// Generate both query variations
		queries := []string{
			track.FullQuery(retry.QuerySuffix),       // 1. All Artists
			track.MainArtistQuery(retry.QuerySuffix), // 2. Primary Artist Only (Fallback)
		}

		for _, searchQuery := range queries {
			matchFilter := fmt.Sprintf("duration >= %d & duration <= %d",
				track.DurationSec-retry.Tolerance,
				track.DurationSec+retry.Tolerance,
			)

			args := []string{
				fmt.Sprintf("ytsearch1:%s", searchQuery),
				"--match-filter", matchFilter,
				"--extract-audio", "--audio-format", "mp3", "--audio-quality", "0",
				"--add-metadata", "-P", (*cfg).DownloadPath,
				"-o", "%(title)s.%(ext)s", "--no-playlist", "--embed-thumbnail", "--verbose",
			}

			cmd := exec.CommandContext(ctx, "yt-dlp", args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Handle Hard Errors
			if err != nil && !strings.Contains(outputStr, "does not pass filter") {
				if (*cfg).Debug {
					fmt.Printf("[ENGINE] DEBUG: Hard Error: %v\n", outputStr)
				}
				continue // Try the primary artist fallback if the first query errored
			}

			// Handle Filter Skips
			if strings.Contains(outputStr, "does not pass filter") {
				// If this was the FullQuery, the loop will automatically try MainArtistQuery next
				continue
			}

			// Handle Success
			if strings.Contains(outputStr, "[download] Destination:") || strings.Contains(outputStr, "has already been downloaded") {
				fmt.Printf("[ENGINE] Success: %s (Tolerance: %ds) via *%v*\n", track.Title, retry.Tolerance, searchQuery)
				return nil
			}
			fmt.Printf("[ENGINE] Attempt %d: No match for %s within ±%ds with query *%v*. Moving to next retry level...\n", i+1, track.Title, retry.Tolerance, searchQuery)
		}

	}

	return fmt.Errorf("failed to download %s after %d levels of retries", track.Title, maxRetries)
}
