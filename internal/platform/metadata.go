package platform

import (
	"GoDown/internal/models"
	"fmt"
	"net/url"
	"strings"
)

// Takes a URL to a playlist or track and returns an slice of Track's
// if there was an error, it returns an empty slice
func FetchMetadata(URL string) ([]models.Track, error) {
	var tracks []models.Track
	var err error

	u, err := url.Parse(URL)
	if err != nil {
		return []models.Track{}, fmt.Errorf("error parsing url: %w", err)
	}

	provider := u.Host
	if provider == "open.spotify.com" {
		pathSegments := strings.Split(strings.Trim(u.Path, "/"), "/")
		url := fmt.Sprintf("https://open.spotify.com/embed/%v/%v", pathSegments[0], pathSegments[1])

		switch pathSegments[0] {
		case "track":
			tracks, err = SpotifyTrackMetadata(url)
		case "playlist", "album":
			tracks, err = SpotifyPlaylistMetadata(url)
		}

		if err != nil {
			return []models.Track{}, err
		}
	}

	return tracks, nil
}
