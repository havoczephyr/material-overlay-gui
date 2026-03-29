package service

import (
	"fmt"
	"strconv"
	"strings"
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

	// Fetch card data from YGOPRODECK
	ygoproCard, err := s.ygoproClient.FetchCardByName(cd.Name)
	if err == nil && ygoproCard != nil {
		// Use YGOPRODECK's plain text description instead of Yugipedia wikitext lore
		if ygoproCard.Desc != "" {
			cd.Lore = ygoproCard.Desc
		}
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

// FetchTips fetches card tips and returns clean plain text.
func (s *CardService) FetchTips(cardName string) (string, error) {
	raw, err := s.wikiClient.FetchCardTips(cardName)
	if err != nil {
		return "", err
	}
	return card.StripWikiPageContent(raw), nil
}

// FetchTrivia fetches card trivia and returns clean plain text.
func (s *CardService) FetchTrivia(cardName string) (string, error) {
	raw, err := s.wikiClient.FetchCardTrivia(cardName)
	if err != nil {
		return "", err
	}
	return card.StripWikiPageContent(raw), nil
}

// FetchRulings fetches card rulings and returns clean plain text.
func (s *CardService) FetchRulings(cardName string) (string, error) {
	raw, err := s.wikiClient.FetchCardRulings(cardName)
	if err != nil {
		return "", err
	}
	return card.StripWikiPageContent(raw), nil
}

// FetchErrata fetches card errata and returns clean plain text.
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

// FetchRecentSets returns the most recent card sets, sorted by TCG date descending.
func (s *CardService) FetchRecentSets(limit int) ([]api.YGOProSet, error) {
	// Check cache
	if cached, ok := s.cache.Get("sets:all"); ok {
		if sets, ok := cached.([]api.YGOProSet); ok {
			return truncateSets(sets, limit), nil
		}
	}

	sets, err := s.ygoproClient.FetchAllSets()
	if err != nil {
		return nil, err
	}

	// Sort by TCG date descending
	sortSetsByDate(sets)

	// Cache for 24h
	s.cache.Set("sets:all", sets, cache.SetTTL)

	return truncateSets(sets, limit), nil
}

// FetchCardsInSet returns all cards in a specific set.
func (s *CardService) FetchCardsInSet(setName string) ([]api.YGOProCard, error) {
	return s.ygoproClient.FetchCardsBySet(setName)
}

// FetchArchetypeArticle fetches the wiki article for an archetype as clean plain text.
// Tries the plain name first, then falls back to "Name (archetype)".
func (s *CardService) FetchArchetypeArticle(name string) (string, error) {
	raw, err := s.wikiClient.FetchWikiPage(name)
	if err != nil {
		return "", err
	}
	if raw != "" {
		return card.StripWikiPageContent(raw), nil
	}
	raw, err = s.wikiClient.FetchWikiPage(name + " (archetype)")
	if err != nil {
		return "", err
	}
	return card.StripWikiPageContent(raw), nil
}

// FetchArchetypeCards fetches all cards belonging to an archetype from YGOPRODECK.
func (s *CardService) FetchArchetypeCards(name string) ([]api.YGOProCard, error) {
	return s.ygoproClient.FetchCardsByArchetype(name)
}

// FetchArchetypeSplashImage fetches the first relevant image from the archetype's wiki page.
func (s *CardService) FetchArchetypeSplashImage(name string) ([]byte, error) {
	// Check disk cache
	cacheKey := "archetype_" + name
	data, err := s.imageCache.GetCachedImageByName(cacheKey)
	if err == nil && data != nil {
		return data, nil
	}

	// Try plain name first, then "(archetype)" variant
	images, _ := s.wikiClient.FetchPageImages(name)
	if len(images) == 0 {
		images, _ = s.wikiClient.FetchPageImages(name + " (archetype)")
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("no images found for archetype %q", name)
	}

	// Pick first image that looks like card art (skip icons/logos)
	var filename string
	for _, img := range images {
		lower := strings.ToLower(img)
		if strings.HasSuffix(lower, ".svg") {
			continue
		}
		if strings.Contains(lower, "icon") || strings.Contains(lower, "logo") {
			continue
		}
		filename = img
		break
	}
	if filename == "" {
		filename = images[0]
	}

	url, err := s.wikiClient.FetchFileImageURL(filename)
	if err != nil {
		return nil, err
	}

	data, err = s.wikiClient.DownloadImage(url)
	if err != nil {
		return nil, err
	}

	_ = s.imageCache.CacheImageByName(cacheKey, data)
	return data, nil
}

// CategorizeRecentSets splits sets into packs and structure decks.
func CategorizeRecentSets(sets []api.YGOProSet) (packs, structures []api.YGOProSet) {
	for _, s := range sets {
		name := strings.ToLower(s.SetName)
		if strings.Contains(name, "structure deck") || strings.Contains(name, "starter deck") {
			structures = append(structures, s)
		} else {
			packs = append(packs, s)
		}
	}
	return
}

func sortSetsByDate(sets []api.YGOProSet) {
	// Sort by TCGDate descending (format: "2024-01-15")
	for i := 1; i < len(sets); i++ {
		for j := i; j > 0 && sets[j].TCGDate > sets[j-1].TCGDate; j-- {
			sets[j], sets[j-1] = sets[j-1], sets[j]
		}
	}
}

func truncateSets(sets []api.YGOProSet, limit int) []api.YGOProSet {
	if len(sets) <= limit {
		return sets
	}
	return sets[:limit]
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
