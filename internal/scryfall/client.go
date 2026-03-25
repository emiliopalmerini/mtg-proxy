package scryfall

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/epalmerini/mtg-proxy/internal/card"
)

type apiResponse struct {
	Object     string  `json:"object"`
	Name       string  `json:"name"`
	ManaCost   string  `json:"mana_cost"`
	TypeLine   string  `json:"type_line"`
	OracleText string  `json:"oracle_text"`
	Power      *string `json:"power,omitempty"`
	Toughness  *string `json:"toughness,omitempty"`
	Loyalty    *string `json:"loyalty,omitempty"`
	Details    string  `json:"details,omitempty"`
}

const requestDelay = 100 * time.Millisecond

// Client fetches card data from the Scryfall API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex
	lastReq    time.Time
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) throttle() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elapsed := time.Since(c.lastReq); elapsed < requestDelay {
		time.Sleep(requestDelay - elapsed)
	}
	c.lastReq = time.Now()
}

func (c *Client) FetchCard(entry card.DeckEntry) (card.Card, error) {
	c.throttle()

	var reqURL string
	if !entry.SetCode.IsEmpty() && !entry.CollectorNumber.IsEmpty() {
		reqURL = fmt.Sprintf("%s/cards/%s/%s",
			c.baseURL,
			url.PathEscape(string(entry.SetCode)),
			url.PathEscape(string(entry.CollectorNumber)),
		)
	} else {
		reqURL = fmt.Sprintf("%s/cards/named?exact=%s",
			c.baseURL,
			url.QueryEscape(string(entry.Name)),
		)
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return card.Card{}, fmt.Errorf("building request for %q: %w", entry.Name, err)
	}
	req.Header.Set("User-Agent", "mtg-proxy/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return card.Card{}, fmt.Errorf("fetching card %q: %w", entry.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp apiResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return card.Card{}, fmt.Errorf("card %q not found: %s", entry.Name, errResp.Details)
	}

	var data apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return card.Card{}, fmt.Errorf("decoding card %q: %w", entry.Name, err)
	}

	result := card.Card{
		Name:       card.CardName(data.Name),
		ManaCost:   card.ParseManaCost(data.ManaCost),
		TypeLine:   card.TypeLine(data.TypeLine),
		OracleText: card.OracleText(data.OracleText),
	}

	if data.Power != nil && data.Toughness != nil {
		result.Stats = &card.Stats{Power: *data.Power, Toughness: *data.Toughness}
	}

	if data.Loyalty != nil {
		l := card.Loyalty(*data.Loyalty)
		result.Loyalty = &l
	}

	return result, nil
}
