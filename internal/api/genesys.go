package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// GenesysPointMap maps card names to their Genesys point costs.
// Cards not in the map have a default cost of 0.
type GenesysPointMap map[string]int

// GenesysData holds the point list name and parsed point map for local caching.
type GenesysData struct {
	ListName string          `json:"list_name"`
	Points   GenesysPointMap `json:"points"`
}

// FindCurrentPointListName returns the name of the most recent Genesys Point List page.
func (c *Client) FindCurrentPointListName() (string, error) {
	params := []Param{
		{"action", "ask"},
		{"query", "[[Category:Genesys Point Lists]]|?Start date|sort=Start date|order=desc|limit=1"},
		{"format", "json"},
	}

	body, err := c.doRequest(params)
	if err != nil {
		return "", err
	}

	var resp SMWResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parsing SMW response: %w", err)
	}

	for name := range resp.Query.Results {
		return name, nil
	}

	return "", fmt.Errorf("no Genesys Point List pages found")
}

// FetchPointList fetches a specific point list page and parses it.
func (c *Client) FetchPointList(pageName string) (GenesysPointMap, error) {
	raw, err := c.fetchPageWikitext(pageName)
	if err != nil {
		return nil, fmt.Errorf("fetching point list source: %w", err)
	}
	return parsePointList(raw), nil
}

// parseResponse is the structure for MediaWiki action=parse responses.
type parseResponse struct {
	Parse struct {
		Wikitext struct {
			Content string `json:"*"`
		} `json:"wikitext"`
	} `json:"parse"`
}

// fetchPageWikitext gets the raw wikitext of a page via action=parse.
func (c *Client) fetchPageWikitext(title string) (string, error) {
	params := []Param{
		{"action", "parse"},
		{"page", title},
		{"prop", "wikitext"},
		{"format", "json"},
	}

	body, err := c.doRequest(params)
	if err != nil {
		return "", err
	}

	var resp parseResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	return resp.Parse.Wikitext.Content, nil
}

// parsePointList extracts card->cost pairs from Genesys point list wikitext.
// Format: CardName; cost (one per line), inside a {{Genesys point list}} template.
func parsePointList(wikitext string) GenesysPointMap {
	points := make(GenesysPointMap)

	for _, line := range strings.Split(wikitext, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "|") || strings.HasPrefix(line, "{") || strings.HasPrefix(line, "}") {
			continue
		}

		parts := strings.SplitN(line, ";", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		if len(name) >= 2 && name[0] == '"' && name[len(name)-1] == '"' {
			name = name[1 : len(name)-1]
		}
		costStr := strings.TrimSpace(parts[1])

		cost, err := strconv.Atoi(costStr)
		if err != nil {
			continue
		}

		points[name] = cost
	}

	return points
}
