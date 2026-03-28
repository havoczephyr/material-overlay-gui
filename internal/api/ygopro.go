package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/havoczephyr/material-overlay-gui/internal/ratelimit"
)

const ygoproBaseURL = "https://db.ygoprodeck.com/api/v7"

// YGOProClient handles requests to the YGOPRODECK API.
// Uses Go's net/http (no Cloudflare issues like Yugipedia).
type YGOProClient struct {
	httpClient  *http.Client
	rateLimiter *ratelimit.Limiter
}

func NewYGOProClient(limiter *ratelimit.Limiter) *YGOProClient {
	return &YGOProClient{
		httpClient:  &http.Client{},
		rateLimiter: limiter,
	}
}

// YGOProCardResponse is the top-level response from cardinfo.php.
type YGOProCardResponse struct {
	Data []YGOProCard `json:"data"`
}

type YGOProCard struct {
	ID        int              `json:"id"`
	Name      string           `json:"name"`
	Type      string           `json:"type"`
	FrameType string           `json:"frameType"`
	Desc      string           `json:"desc"`
	ATK       int              `json:"atk"`
	DEF       int              `json:"def"`
	Level     int              `json:"level"`
	Race      string           `json:"race"`
	Attribute string           `json:"attribute"`
	CardSets  []YGOProCardSet  `json:"card_sets"`
	CardImages []YGOProImage   `json:"card_images"`
}

type YGOProCardSet struct {
	SetName   string `json:"set_name"`
	SetCode   string `json:"set_code"`
	SetRarity string `json:"set_rarity"`
	SetPrice  string `json:"set_price"`
}

type YGOProImage struct {
	ID              int    `json:"id"`
	ImageURL        string `json:"image_url"`
	ImageURLSmall   string `json:"image_url_small"`
	ImageURLCropped string `json:"image_url_cropped"`
}

type YGOProSet struct {
	SetName    string `json:"set_name"`
	SetCode    string `json:"set_code"`
	NumOfCards int    `json:"num_of_cards"`
	TCGDate    string `json:"tcg_date"`
}

// FetchRandomCard returns a random card from the full TCG/OCG pool.
func (c *YGOProClient) FetchRandomCard() (*YGOProCard, error) {
	c.rateLimiter.Wait()

	resp, err := c.httpClient.Get(ygoproBaseURL + "/randomcard.php")
	if err != nil {
		return nil, fmt.Errorf("fetching random card: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("random card API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading random card response: %w", err)
	}

	var cardResp YGOProCardResponse
	if err := json.Unmarshal(body, &cardResp); err != nil {
		return nil, fmt.Errorf("parsing random card: %w", err)
	}

	if len(cardResp.Data) == 0 {
		return nil, fmt.Errorf("random card API returned no data")
	}

	return &cardResp.Data[0], nil
}

// FetchCardByName returns card data including image URLs and set membership.
func (c *YGOProClient) FetchCardByName(name string) (*YGOProCard, error) {
	c.rateLimiter.Wait()

	req, err := http.NewRequest("GET", ygoproBaseURL+"/cardinfo.php", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("name", name)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching card %q: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("card API returned status %d for %q", resp.StatusCode, name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading card response: %w", err)
	}

	var cardResp YGOProCardResponse
	if err := json.Unmarshal(body, &cardResp); err != nil {
		return nil, fmt.Errorf("parsing card response for %q: %w", name, err)
	}

	if len(cardResp.Data) == 0 {
		return nil, fmt.Errorf("no card found for %q", name)
	}

	return &cardResp.Data[0], nil
}

// FetchCardImage downloads a card image from the given URL.
// Returns the raw image bytes. Caller is responsible for caching to disk per YGOPRODECK ToS.
func (c *YGOProClient) FetchCardImage(imageURL string) ([]byte, error) {
	c.rateLimiter.Wait()

	resp, err := c.httpClient.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("fetching card image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image request returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading card image: %w", err)
	}

	return data, nil
}

// FetchAllSets returns all card sets from YGOPRODECK.
func (c *YGOProClient) FetchAllSets() ([]YGOProSet, error) {
	c.rateLimiter.Wait()

	resp, err := c.httpClient.Get(ygoproBaseURL + "/cardsets.php")
	if err != nil {
		return nil, fmt.Errorf("fetching card sets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cardsets API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading sets response: %w", err)
	}

	var sets []YGOProSet
	if err := json.Unmarshal(body, &sets); err != nil {
		return nil, fmt.Errorf("parsing sets response: %w", err)
	}

	return sets, nil
}

// FetchCardsBySet returns all cards in a specific set from YGOPRODECK.
func (c *YGOProClient) FetchCardsBySet(setName string) ([]YGOProCard, error) {
	c.rateLimiter.Wait()

	req, err := http.NewRequest("GET", ygoproBaseURL+"/cardinfo.php", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("cardset", setName)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching cards in set %q: %w", setName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("card API returned status %d for set %q", resp.StatusCode, setName)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading set cards response: %w", err)
	}

	var cardResp YGOProCardResponse
	if err := json.Unmarshal(body, &cardResp); err != nil {
		return nil, fmt.Errorf("parsing set cards for %q: %w", setName, err)
	}

	return cardResp.Data, nil
}
