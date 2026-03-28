package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

var cardProperties = []string{
	"Card type",
	"Attribute",
	"ATK string",
	"DEF string",
	"Level string",
	"Rank string",
	"Pendulum Scale string",
	"Link Arrows",
	"Types",
	"Lore",
	"Property",
	"Archseries",
	"Password",
	"TCG status",
	"OCG status",
}

func (c *Client) LookupCard(name string) (*SMWResponse, error) {
	props := make([]string, len(cardProperties))
	for i, p := range cardProperties {
		props[i] = "?" + p
	}
	query := "[[" + name + "]]|" + strings.Join(props, "|")

	params := []Param{
		{"action", "ask"},
		{"query", query},
		{"format", "json"},
	}

	body, err := c.doRequest(params)
	if err != nil {
		return nil, fmt.Errorf("looking up card %q: %w", name, err)
	}

	var resp SMWResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing SMW response for %q: %w", name, err)
	}

	return &resp, nil
}
