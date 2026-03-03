package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Song struct {
	Title       string
	Artists     []string
	DurationSec int
}

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

func main() {
	fmt.Println("Hello world")

	playlistSongs := []Song{}

	// Hard coded ID for now
	playlistID := "0uuSKJdtUWEmgsGHvQbs5O"
	url := fmt.Sprintf("https://open.spotify.com/embed/playlist/%s", playlistID)

	c := colly.NewCollector(
		colly.AllowedDomains("open.spotify.com"),
	)

	// htmlElm := "ol[aria-label=Track list]"
	c.OnHTML("li", func(h *colly.HTMLElement) {
		var currentSong Song

		title := h.ChildText("h3")

		artistRaw := ""
		h.DOM.Find("h4").Contents().Each(func(i int, s *goquery.Selection) {
			if goquery.NodeName(s) == "#text" {
				artistRaw += s.Text()
			}
		})
		artists := strings.Split(artistRaw, "\u00a0")
		for i, a := range artists {
			// removes trailing comma for every artist except last
			if i != (len(artists) - 1) {
				artists[i] = a[:len(a)-1]
			} else {
				artists[i] = strings.TrimSpace(a)
			}
		}

		duration, err := toSec(h.ChildText("div[data-testid=duration-cell]"))
		if err != nil {
			fmt.Printf("error getting duration: %v", err)
		}

		if title != "" && len(artists) != 0 && duration != -1 {
			fmt.Printf("Title: %s\nMain Artist: *%v*\nDuration: %v\n\n", title, artists, duration)
			currentSong.Title = title
			currentSong.Artists = artists
			currentSong.DurationSec = duration

			playlistSongs = append(playlistSongs, currentSong)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(url)
}
