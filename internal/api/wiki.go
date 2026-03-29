package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ParseResponse is the top-level response from action=parse with prop=wikitext.
type ParseResponse struct {
	Parse ParseResult `json:"parse"`
}

// ParseResult holds the parsed page data.
type ParseResult struct {
	Title    string        `json:"title"`
	PageID   int           `json:"pageid"`
	Wikitext ParseWikitext `json:"wikitext"`
}

// ParseWikitext holds the raw wikitext content.
type ParseWikitext struct {
	Content string `json:"*"`
}

// FetchWikiPage retrieves the raw wikitext content of a Yugipedia page by title.
// Returns empty string and nil error if the page doesn't exist.
func (c *Client) FetchWikiPage(title string) (string, error) {
	params := []Param{
		{"action", "parse"},
		{"page", title},
		{"prop", "wikitext"},
		{"format", "json"},
	}

	body, err := c.doRequest(params)
	if err != nil {
		return "", nil
	}

	var resp ParseResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", nil
	}

	return resp.Parse.Wikitext.Content, nil
}

// FetchCardTips fetches the "Card Tips:<name>" page from Yugipedia.
func (c *Client) FetchCardTips(cardName string) (string, error) {
	return c.FetchWikiPage("Card Tips:" + cardName)
}

// FetchCardTrivia fetches the "Card Trivia:<name>" page from Yugipedia.
func (c *Client) FetchCardTrivia(cardName string) (string, error) {
	return c.FetchWikiPage("Card Trivia:" + cardName)
}

// FetchCardRulings fetches the "Card Rulings:<name>" page from Yugipedia.
func (c *Client) FetchCardRulings(cardName string) (string, error) {
	return c.FetchWikiPage("Card Rulings:" + cardName)
}

// FetchCardErrata fetches the "Card Errata:<name>" page from Yugipedia.
func (c *Client) FetchCardErrata(cardName string) (string, error) {
	return c.FetchWikiPage("Card Errata:" + cardName)
}

// ── Gallery image methods ───────────────────────────────────────────────

// ParseImagesResponse holds the response from action=parse with prop=images.
type ParseImagesResponse struct {
	Parse struct {
		Images []string `json:"images"`
	} `json:"parse"`
}

// ImageInfoResponse holds the response from action=query with prop=imageinfo.
type ImageInfoResponse struct {
	Query struct {
		Pages map[string]ImageInfoPage `json:"pages"`
	} `json:"query"`
}

// ImageInfoPage is a single page entry in the imageinfo response.
type ImageInfoPage struct {
	PageID    int             `json:"pageid"`
	Title     string          `json:"title"`
	ImageInfo []ImageInfoItem `json:"imageinfo"`
}

// ImageInfoItem holds a single image's metadata.
type ImageInfoItem struct {
	URL string `json:"url"`
}

// FetchPageImages returns all image filenames from any Yugipedia page.
// Returns nil if the page doesn't exist.
func (c *Client) FetchPageImages(title string) ([]string, error) {
	params := []Param{
		{"action", "parse"},
		{"page", title},
		{"prop", "images"},
		{"format", "json"},
	}

	body, err := c.doRequest(params)
	if err != nil {
		return nil, nil
	}

	var resp ParseImagesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil
	}

	return resp.Parse.Images, nil
}

// FetchGalleryImageNames returns all image filenames from a card's gallery page.
// Returns nil if the gallery page doesn't exist.
func (c *Client) FetchGalleryImageNames(cardName string) ([]string, error) {
	return c.FetchPageImages("Card Gallery:" + cardName)
}

// FetchFileImageURL resolves a single Yugipedia filename to its direct download URL.
func (c *Client) FetchFileImageURL(filename string) (string, error) {
	params := []Param{
		{"action", "query"},
		{"titles", "File:" + filename},
		{"prop", "imageinfo"},
		{"iiprop", "url"},
		{"format", "json"},
	}

	body, err := c.doRequest(params)
	if err != nil {
		return "", fmt.Errorf("fetching file URL for %q: %w", filename, err)
	}

	var resp ImageInfoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parsing imageinfo for %q: %w", filename, err)
	}

	for _, page := range resp.Query.Pages {
		if len(page.ImageInfo) > 0 {
			return page.ImageInfo[0].URL, nil
		}
	}

	return "", fmt.Errorf("no image URL found for %q", filename)
}

// DownloadImage fetches raw image bytes from a direct URL via curl.
// Does NOT go through the API rate limiter (media CDN is separate).
func (c *Client) DownloadImage(imageURL string) ([]byte, error) {
	cmd := exec.Command("curl",
		"-s",
		"--globoff",
		"-f",
		"--compressed",
		"-L", // follow redirects
		"-H", "User-Agent: "+c.userAgent,
		imageURL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("downloading image: %w: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// GalleryEntry represents a single image in a card gallery.
type GalleryEntry struct {
	Filename string // e.g., "DarkMagician-LOB-EN-UR-UE.png"
	Label    string // e.g., "LOB EN UR UE"
}

// ParseGalleryEntries filters and parses gallery image filenames for a given card.
// Returns only filenames that match the card's naming pattern.
func ParseGalleryEntries(cardName string, filenames []string) []GalleryEntry {
	prefix := cardFilePrefix(cardName)
	var entries []GalleryEntry

	for _, f := range filenames {
		if !strings.HasPrefix(f, prefix+"-") {
			continue
		}
		// Skip cropped/small variants
		lower := strings.ToLower(f)
		if strings.Contains(lower, "-crop") || strings.Contains(lower, "-small") {
			continue
		}
		label := parseFilenameLabel(f, prefix)
		entries = append(entries, GalleryEntry{Filename: f, Label: label})
	}

	return entries
}

// cardFilePrefix generates the filename prefix from a card name.
// "Dark Magician" -> "DarkMagician", "Mekk-Knight Green Horizon" -> "MekkKnightGreenHorizon"
func cardFilePrefix(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// parseFilenameLabel extracts set/rarity/edition info from a gallery filename.
// "DarkMagician-LOB-EN-UR-UE.png" -> "LOB EN UR UE"
func parseFilenameLabel(filename, prefix string) string {
	// Strip prefix and extension
	s := strings.TrimPrefix(filename, prefix+"-")
	if idx := strings.LastIndex(s, "."); idx != -1 {
		s = s[:idx]
	}
	// Replace hyphens with spaces for readability
	return strings.ReplaceAll(s, "-", " ")
}
