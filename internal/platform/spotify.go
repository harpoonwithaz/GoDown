package platform

import (
	"fmt"
	"net/url"
	"path"
	"spotify-playlist-downloader/internal/models"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

// Converts string minute:second format to second int
// Returns -1 and err if it was unsuccessful
func toSec(str string) (int, error) {
	lengthSeconds := 0

	delimPos := len(str) - 2
	sec, err := strconv.Atoi(str[delimPos:])

	if err != nil {
		return -1, err
	} else {
		lengthSeconds += sec
	}

	min, err := strconv.Atoi(str[:delimPos-1]) // not including the delim
	if err != nil {
		return -1, err
	} else {
		lengthSeconds += (min * 60)
	}

	return lengthSeconds, nil
}

// Returns the playlist items from a spotify URL as a track struct
func FetchMetadata(URL string) ([]models.Track, error) {
	playlistTracks := []models.Track{}

	// Parse the url for playlist ID
	u, err := url.Parse(URL)
	if err != nil {
		return []models.Track{}, fmt.Errorf("error parsing url: %w", err)
	}
	playlistID := path.Base(u.Path)
	scrapeURL := fmt.Sprintf("https://open.spotify.com/embed/playlist/%s", playlistID)

	// scrape the embed HTML for track info
	c := colly.NewCollector(
		colly.AllowedDomains("open.spotify.com"),
	)

	var visitErr error

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			visitErr = fmt.Errorf("unexpected status code: %d", r.StatusCode)
			return
		}
		body := strings.ToLower(string(r.Body))
		if strings.Contains(body, "page not found") || strings.Contains(body, "this content is not available") || strings.Contains(body, "playlist is private") {
			visitErr = fmt.Errorf("playlist not available or private")
			return
		}
	})

	c.OnError(func(r *colly.Response, er error) {
		// network / request error
		visitErr = fmt.Errorf("request failed: %w", er)
	})

	// htmlElm := "ol[aria-label=Track list]"
	c.OnHTML("li", func(h *colly.HTMLElement) {
		var currentTrack models.Track

		title := h.ChildText("h3")

		// Ignore span tag when track is explicit
		artistRaw := ""
		h.DOM.Find("h4").Contents().Each(func(i int, s *goquery.Selection) {
			if goquery.NodeName(s) == "#text" {
				artistRaw += s.Text()
			}
		})

		// Multiple artists
		artists := strings.Split(artistRaw, "\u00a0")
		for i, a := range artists {
			// removes trailing comma for every artist except last
			if i != (len(artists) - 1) {
				artists[i] = a[:len(a)-1]
			} else {
				artists[i] = strings.TrimSpace(a)
			}
		}

		// convert to seconds from minutes:seconds format
		durationText := h.ChildText("div[data-testid=duration-cell]")
		duration, err := toSec(durationText)
		if err != nil {
			fmt.Printf("error converting duration: %v\nDuration text %v\nDuration set to -1", err, durationText)
		}

		if title != "" && len(artists) != 0 {
			// fmt.Printf("Title: %s\nMain Artist: *%v*\nDuration: %v\n\n", title, artists, duration)
			currentTrack.Title = title
			currentTrack.Artists = artists
			currentTrack.DurationSec = duration

			playlistTracks = append(playlistTracks, currentTrack)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(scrapeURL)

	if visitErr != nil {
		return []models.Track{}, visitErr
	}

	return playlistTracks, nil
}
