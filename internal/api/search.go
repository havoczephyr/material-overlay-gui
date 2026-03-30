package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// gameSuffixes are parenthetical tags for non-TCG/OCG card pages to filter out.
var gameSuffixes = []string{
	"(Master Duel)",
	"(Duel Links)",
	"(Rush Duel)",
	"(Tag Force)",
	"(BAM)",
	"(Duel Arena)",
	"(Legacy of the Duelist)",
	"(World Championship)",
	"(GX anime)",
	"(anime)",
	"(manga)",
	"(character)",
	"(series)",
	"(archetype)",
}

func (c *Client) SearchCards(query string, limit int) ([]SearchResult, error) {
	params := []Param{
		{"action", "query"},
		{"list", "search"},
		{"srsearch", query},
		{"srnamespace", "0"},
		{"srlimit", fmt.Sprintf("%d", limit)},
		{"format", "json"},
	}

	body, err := c.doRequest(params)
	if err != nil {
		return nil, fmt.Errorf("searching for %q: %w", query, err)
	}

	var resp SearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing search response for %q: %w", query, err)
	}

	return filterResults(resp.Query.Search), nil
}

// SearchWikiPages searches namespace 0 without filtering game-specific suffixes.
func (c *Client) SearchWikiPages(query string, limit int) ([]SearchResult, error) {
	params := []Param{
		{"action", "query"},
		{"list", "search"},
		{"srsearch", query},
		{"srnamespace", "0"},
		{"srlimit", fmt.Sprintf("%d", limit)},
		{"format", "json"},
	}

	body, err := c.doRequest(params)
	if err != nil {
		return nil, fmt.Errorf("searching wiki pages for %q: %w", query, err)
	}

	var resp SearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing wiki search response for %q: %w", query, err)
	}

	return resp.Query.Search, nil
}

// HasGameSuffix reports whether a title ends with a known game-specific suffix.
func HasGameSuffix(title string) bool {
	return hasGameSuffix(title)
}

func filterResults(results []SearchResult) []SearchResult {
	var filtered []SearchResult
	for _, r := range results {
		if !hasGameSuffix(r.Title) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func hasGameSuffix(title string) bool {
	for _, suffix := range gameSuffixes {
		if strings.HasSuffix(title, suffix) {
			return true
		}
	}
	return false
}
