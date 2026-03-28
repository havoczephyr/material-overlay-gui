package service

import (
	"encoding/json"
	"os"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
)

const genesysCachePath = ".cache/genesys.json"

// GenesysService handles loading and caching the Genesys point list.
type GenesysService struct {
	client *api.Client
	Points api.GenesysPointMap
}

func NewGenesysService(client *api.Client) *GenesysService {
	return &GenesysService{client: client}
}

// Load attempts to load points from disk cache first, then from wiki.
func (g *GenesysService) Load() error {
	// Try disk cache first
	if data, err := g.loadFromDisk(); err == nil && data != nil {
		g.Points = data.Points
		return nil
	}

	// Fetch from wiki
	listName, err := g.client.FindCurrentPointListName()
	if err != nil {
		return err
	}

	points, err := g.client.FetchPointList(listName)
	if err != nil {
		return err
	}

	g.Points = points

	// Save to disk
	_ = g.saveToDisk(&api.GenesysData{
		ListName: listName,
		Points:   points,
	})

	return nil
}

func (g *GenesysService) loadFromDisk() (*api.GenesysData, error) {
	data, err := os.ReadFile(genesysCachePath)
	if err != nil {
		return nil, err
	}

	var gd api.GenesysData
	if err := json.Unmarshal(data, &gd); err != nil {
		return nil, err
	}

	return &gd, nil
}

func (g *GenesysService) saveToDisk(gd *api.GenesysData) error {
	_ = os.MkdirAll(".cache", 0o755)
	data, err := json.Marshal(gd)
	if err != nil {
		return err
	}
	return os.WriteFile(genesysCachePath, data, 0o644)
}
