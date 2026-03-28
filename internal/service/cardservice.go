package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
	"github.com/havoczephyr/material-overlay-gui/internal/cache"
	"github.com/havoczephyr/material-overlay-gui/internal/card"
	"github.com/havoczephyr/material-overlay-gui/internal/imagecache"
	"github.com/havoczephyr/material-overlay-gui/internal/ratelimit"
)

// CardService orchestrates API calls, caching, and image fetching.
type CardService struct {
	wikiClient   *api.Client
	ygoproClient *api.YGOProClient
	cache        *cache.Cache
	imageCache   *imagecache.DiskCache
	genesys      *GenesysService
}

func NewCardService() (*CardService, error) {
	imgCache, err := imagecache.New()
	if err != nil {
		return nil, err
	}

	wikiLimiter := ratelimit.New()
	ygoproLimiter := ratelimit.NewWithInterval(50 * time.Millisecond)

	wikiClient := api.NewClient(wikiLimiter)
	ygoproClient := api.NewYGOProClient(ygoproLimiter)
	memCache := cache.New()
	genesys := NewGenesysService(wikiClient)

	return &CardService{
		wikiClient:   wikiClient,
		ygoproClient: ygoproClient,
		cache:        memCache,
		imageCache:   imgCache,
		genesys:      genesys,
	}, nil
}

// LookupCard fetches card data from Yugipedia and returns the card with image bytes.
func (s *CardService) LookupCard(name string) (*card.Card, []byte, error) {
	// Check memory cache
	if cd, ok := s.cache.GetCard(name); ok {
		imgData := s.fetchImageForCard(cd)
		return cd, imgData, nil
	}

	// SMW lookup
	resp, err := s.wikiClient.LookupCard(name)
	if err != nil {
		return nil, nil, fmt.Errorf("looking up card: %w", err)
	}

	if len(resp.Query.Results) == 0 {
		return nil, nil, fmt.Errorf("card %q not found", name)
	}

	var entry api.SMWResultEntry
	for _, e := range resp.Query.Results {
		entry = e
		break
	}

	cd := card.NewCardFromSMW(name, entry)

	// Apply Genesys cost
	if s.genesys.Points != nil {
		if cost, ok := s.genesys.Points[cd.Name]; ok {
			cd.GenesysCost = strconv.Itoa(cost)
		} else {
			cd.GenesysCost = "0"
		}
	}

	// Fetch card sets from YGOPRODECK
	ygoproCard, err := s.ygoproClient.FetchCardByName(cd.Name)
	if err == nil && ygoproCard != nil {
		for _, cs := range ygoproCard.CardSets {
			cd.CardSets = append(cd.CardSets, card.CardSetEntry{
				SetName: cs.SetName,
				SetCode: cs.SetCode,
				Rarity:  cs.SetRarity,
			})
		}
		if len(ygoproCard.CardImages) > 0 {
			cd.ImageURL = ygoproCard.CardImages[0].ImageURL
		}
	}

	// Cache card
	s.cache.SetCard(name, &cd)

	// Fetch image
	imgData := s.fetchImageForCard(&cd)

	return &cd, imgData, nil
}

// SearchCards searches Yugipedia for cards matching the query.
func (s *CardService) SearchCards(query string) ([]api.SearchResult, error) {
	return s.wikiClient.SearchCards(query, 20)
}

// LoadRandomCard fetches a random card from YGOPRODECK.
func (s *CardService) LoadRandomCard() (*api.YGOProCard, []byte, error) {
	ygoproCard, err := s.ygoproClient.FetchRandomCard()
	if err != nil {
		return nil, nil, err
	}

	var imgData []byte
	if len(ygoproCard.CardImages) > 0 {
		imgData = s.fetchImageByID(ygoproCard.CardImages[0].ID, ygoproCard.CardImages[0].ImageURL)
	}

	return ygoproCard, imgData, nil
}

// FetchTips fetches card tips from Yugipedia.
func (s *CardService) FetchTips(cardName string) (string, error) {
	raw, err := s.wikiClient.FetchCardTips(cardName)
	if err != nil {
		return "", err
	}
	return card.StripWikiPageContent(raw), nil
}

// FetchTrivia fetches card trivia from Yugipedia.
func (s *CardService) FetchTrivia(cardName string) (string, error) {
	raw, err := s.wikiClient.FetchCardTrivia(cardName)
	if err != nil {
		return "", err
	}
	return card.StripWikiPageContent(raw), nil
}

// FetchRulings fetches card rulings from Yugipedia.
func (s *CardService) FetchRulings(cardName string) (string, error) {
	raw, err := s.wikiClient.FetchCardRulings(cardName)
	if err != nil {
		return "", err
	}
	return card.StripWikiPageContent(raw), nil
}

// FetchErrata fetches card errata from Yugipedia.
func (s *CardService) FetchErrata(cardName string) (string, error) {
	raw, err := s.wikiClient.FetchCardErrata(cardName)
	if err != nil {
		return "", err
	}
	return card.StripWikiPageContent(raw), nil
}

// FetchGalleryEntries fetches gallery image entries for a card.
func (s *CardService) FetchGalleryEntries(cardName string) ([]api.GalleryEntry, error) {
	filenames, err := s.wikiClient.FetchGalleryImageNames(cardName)
	if err != nil {
		return nil, err
	}
	return api.ParseGalleryEntries(cardName, filenames), nil
}

// FetchGalleryImage downloads a gallery image by entry.
func (s *CardService) FetchGalleryImage(entry api.GalleryEntry) ([]byte, error) {
	// Check disk cache
	data, err := s.imageCache.GetCachedImageByName(entry.Filename)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return data, nil
	}

	// Resolve URL and download
	url, err := s.wikiClient.FetchFileImageURL(entry.Filename)
	if err != nil {
		return nil, err
	}

	data, err = s.wikiClient.DownloadImage(url)
	if err != nil {
		return nil, err
	}

	// Cache to disk
	_ = s.imageCache.CacheImageByName(entry.Filename, data)

	return data, nil
}

// LoadGenesysPoints loads Genesys point data (from disk cache or wiki).
func (s *CardService) LoadGenesysPoints() error {
	return s.genesys.Load()
}

func (s *CardService) fetchImageForCard(cd *card.Card) []byte {
	if cd.ImageURL == "" {
		// Try YGOPRODECK lookup
		ygoproCard, err := s.ygoproClient.FetchCardByName(cd.Name)
		if err != nil || len(ygoproCard.CardImages) == 0 {
			return nil
		}
		cd.ImageURL = ygoproCard.CardImages[0].ImageURL
		return s.fetchImageByID(ygoproCard.CardImages[0].ID, ygoproCard.CardImages[0].ImageURL)
	}

	// Try to extract card ID from YGOPRODECK
	ygoproCard, err := s.ygoproClient.FetchCardByName(cd.Name)
	if err != nil || len(ygoproCard.CardImages) == 0 {
		return nil
	}
	return s.fetchImageByID(ygoproCard.CardImages[0].ID, cd.ImageURL)
}

func (s *CardService) fetchImageByID(cardID int, imageURL string) []byte {
	// Check disk cache
	data, err := s.imageCache.GetCachedImage(cardID)
	if err == nil && data != nil {
		return data
	}

	// Download
	data, err = s.ygoproClient.FetchCardImage(imageURL)
	if err != nil {
		return nil
	}

	// Cache to disk
	_ = s.imageCache.CacheImage(cardID, data)

	return data
}
