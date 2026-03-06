package downloader

import (
	"context"
	"fmt"
	"os/exec"
	"spotify-playlist-downloader/internal/models"
	"strings"
)

// Download individual track using yt-dlp
func DownloadTrack(ctx context.Context, track models.Track, cfg *models.Config) error {
	maxRetries := len((*cfg).Retries)

	for i, retry := range (*cfg).Retries {
		searchQuery := track.FullQuery(retry.QuerySuffix)

		matchFilter := fmt.Sprintf("duration >= %d & duration <= %d",
			track.DurationSec-retry.Tolerance,
			track.DurationSec+retry.Tolerance,
		)

		args := []string{
			fmt.Sprintf("ytsearch1:%s", searchQuery),
			"--match-filter", matchFilter,
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--add-metadata",
			"-P", (*cfg).DownloadPath,
			"-o", "%(title)s.%(ext)s",
			"--no-playlist",
			"--embed-thumbnail",
		}

		cmd := exec.CommandContext(ctx, "yt-dlp", args...)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Handle Hard Errors (Network down, binary missing, etc.)
		if err != nil && !strings.Contains(outputStr, "does not pass filter") {
			if (*cfg).Debug {
				fmt.Printf("[ENGINE] DEBUG: Hard Error on attempt %d: %v\n", i+1, outputStr)
			}
			// If it's the last attempt, return the error. Otherwise, try next fallback.
			if i+1 == maxRetries {
				return fmt.Errorf("yt-dlp final attempt failed: %w", err)
			}
			continue
		}

		// Handle Filter Skips
		if strings.Contains(outputStr, "does not pass filter") {
			fmt.Printf("[ENGINE] Attempt %d: No match within ±%ds. Widening...\n", i+1, retry.Tolerance)
			continue
		}

		// Handle Success
		if strings.Contains(outputStr, "[download] Destination:") || strings.Contains(outputStr, "has already been downloaded") {
			fmt.Printf("[ENGINE] Success downloading: %s (Tolerance: %ds)\n", track.Title, retry.Tolerance)
			return nil
		}
	}

	return fmt.Errorf("failed to download %s after %d attempts", track.Title, maxRetries)
}
