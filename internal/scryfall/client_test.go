package scryfall_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/epalmerini/mtg-proxy/internal/scryfall"
)

type scryfallResponse struct {
	Object     string  `json:"object"`
	Name       string  `json:"name"`
	ManaCost   string  `json:"mana_cost"`
	TypeLine   string  `json:"type_line"`
	OracleText string  `json:"oracle_text"`
	Power      *string `json:"power,omitempty"`
	Toughness  *string `json:"toughness,omitempty"`
	Loyalty    *string `json:"loyalty,omitempty"`
}

func ptr(s string) *string { return &s }

func newTestServer(t *testing.T, cards map[string]scryfallResponse) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("exact")
		c, ok := cards[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"object":  "error",
				"details": fmt.Sprintf("Card not found: %s", name),
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	}))
}

func TestFetchInstant(t *testing.T) {
	server := newTestServer(t, map[string]scryfallResponse{
		"Lightning Bolt": {
			Object:     "card",
			Name:       "Lightning Bolt",
			ManaCost:   "{R}",
			TypeLine:   "Instant",
			OracleText: "Lightning Bolt deals 3 damage to any target.",
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	c, err := client.FetchCard("Lightning Bolt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Name != "Lightning Bolt" {
		t.Errorf("name: got %q, want %q", c.Name, "Lightning Bolt")
	}
	if c.ManaCost.String() != "R" {
		t.Errorf("mana cost: got %q, want %q", c.ManaCost, "R")
	}
	if c.TypeLine != "Instant" {
		t.Errorf("type line: got %q, want %q", c.TypeLine, "Instant")
	}
	if c.OracleText != "Lightning Bolt deals 3 damage to any target." {
		t.Errorf("oracle text: got %q", c.OracleText)
	}
	if c.Stats != nil {
		t.Errorf("stats: expected nil for instant, got %v", c.Stats)
	}
	if c.Loyalty != nil {
		t.Errorf("loyalty: expected nil for instant, got %v", c.Loyalty)
	}
}

func TestFetchCreature(t *testing.T) {
	server := newTestServer(t, map[string]scryfallResponse{
		"Tarmogoyf": {
			Object:     "card",
			Name:       "Tarmogoyf",
			ManaCost:   "{1}{G}",
			TypeLine:   "Creature — Lhurgoyf",
			OracleText: "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.",
			Power:      ptr("*"),
			Toughness:  ptr("1+*"),
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	c, err := client.FetchCard("Tarmogoyf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Stats == nil {
		t.Fatal("stats: expected non-nil for creature")
	}
	if c.Stats.Power != "*" {
		t.Errorf("power: got %q, want %q", c.Stats.Power, "*")
	}
	if c.Stats.Toughness != "1+*" {
		t.Errorf("toughness: got %q, want %q", c.Stats.Toughness, "1+*")
	}
	if c.Loyalty != nil {
		t.Errorf("loyalty: expected nil for creature, got %v", c.Loyalty)
	}
}

func TestFetchPlaneswalker(t *testing.T) {
	server := newTestServer(t, map[string]scryfallResponse{
		"Jace, the Mind Sculptor": {
			Object:   "card",
			Name:     "Jace, the Mind Sculptor",
			ManaCost: "{2}{U}{U}",
			TypeLine: "Legendary Planeswalker — Jace",
			OracleText: "+2: Look at the top card of target player's library.\n" +
				"0: Draw three cards, then put two cards from your hand on top of your library.\n" +
				"−1: Return target creature to its owner's hand.\n" +
				"−12: Exile all cards from target player's library.",
			Loyalty: ptr("3"),
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	c, err := client.FetchCard("Jace, the Mind Sculptor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Loyalty == nil {
		t.Fatal("loyalty: expected non-nil for planeswalker")
	}
	if *c.Loyalty != "3" {
		t.Errorf("loyalty: got %q, want %q", *c.Loyalty, "3")
	}
	if c.ManaCost.String() != "2UU" {
		t.Errorf("mana cost: got %q, want %q", c.ManaCost, "2UU")
	}
	if c.Stats != nil {
		t.Errorf("stats: expected nil for planeswalker, got %v", c.Stats)
	}
}

func TestFetchNotFound(t *testing.T) {
	server := newTestServer(t, map[string]scryfallResponse{})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	_, err := client.FetchCard("Nonexistent Card")
	if err == nil {
		t.Fatal("expected error for unknown card, got nil")
	}
}

func TestFetchSendsCorrectRequest(t *testing.T) {
	var receivedPath string
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.RequestURI()
		receivedHeaders = r.Header
		json.NewEncoder(w).Encode(scryfallResponse{
			Object: "card", Name: "Lightning Bolt", ManaCost: "{R}", TypeLine: "Instant",
		})
	}))
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	client.FetchCard("Lightning Bolt")

	expected := "/cards/named?exact=Lightning+Bolt"
	if receivedPath != expected {
		t.Errorf("request path: got %q, want %q", receivedPath, expected)
	}
	if ua := receivedHeaders.Get("User-Agent"); ua != "mtg-proxy/1.0" {
		t.Errorf("User-Agent: got %q, want %q", ua, "mtg-proxy/1.0")
	}
	if accept := receivedHeaders.Get("Accept"); accept != "application/json" {
		t.Errorf("Accept: got %q, want %q", accept, "application/json")
	}
}

func TestFetchLand(t *testing.T) {
	server := newTestServer(t, map[string]scryfallResponse{
		"Island": {
			Object:     "card",
			Name:       "Island",
			ManaCost:   "",
			TypeLine:   "Basic Land — Island",
			OracleText: "({T}: Add {U}.)",
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	c, err := client.FetchCard("Island")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !c.ManaCost.IsEmpty() {
		t.Errorf("mana cost: expected empty for land, got %q", c.ManaCost)
	}
	if c.Stats != nil {
		t.Errorf("stats: expected nil for land")
	}
	if c.Loyalty != nil {
		t.Errorf("loyalty: expected nil for land")
	}
}

func TestFetchThrottlesRequests(t *testing.T) {
	server := newTestServer(t, map[string]scryfallResponse{
		"Lightning Bolt": {
			Object: "card", Name: "Lightning Bolt", ManaCost: "{R}", TypeLine: "Instant",
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)

	// First request should be immediate
	start := time.Now()
	client.FetchCard("Lightning Bolt")
	firstDuration := time.Since(start)

	// Second request should be throttled (~100ms delay)
	start = time.Now()
	client.FetchCard("Lightning Bolt")
	secondDuration := time.Since(start)

	if secondDuration < 80*time.Millisecond {
		t.Errorf("second request was too fast (%v), expected >= 80ms throttle", secondDuration)
	}
	// First request should be fast (no throttle needed)
	if firstDuration > 50*time.Millisecond {
		t.Errorf("first request was unexpectedly slow (%v)", firstDuration)
	}
}
