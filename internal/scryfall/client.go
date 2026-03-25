package scryfall

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

// Client fetches card data from the Scryfall API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) FetchCard(name card.CardName) (card.Card, error) {
	reqURL := fmt.Sprintf("%s/cards/named?exact=%s", c.baseURL, url.QueryEscape(string(name)))

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return card.Card{}, fmt.Errorf("building request for %q: %w", name, err)
	}
	req.Header.Set("User-Agent", "mtg-proxy/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return card.Card{}, fmt.Errorf("fetching card %q: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp apiResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return card.Card{}, fmt.Errorf("card %q not found: %s", name, errResp.Details)
	}

	var data apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return card.Card{}, fmt.Errorf("decoding card %q: %w", name, err)
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
