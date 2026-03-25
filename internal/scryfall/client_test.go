package scryfall_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/epalmerini/mtg-proxy/internal/card"
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

func byName(name string) card.DeckEntry {
	return card.DeckEntry{Name: card.CardName(name), Quantity: 1}
}

func bySet(name, set, num string) card.DeckEntry {
	return card.DeckEntry{
		Name:            card.CardName(name),
		Quantity:        1,
		SetCode:         card.SetCode(set),
		CollectorNumber: card.CollectorNumber(num),
	}
}

// newNameServer handles /cards/named?exact=... lookups
func newNameServer(t *testing.T, cards map[string]scryfallResponse) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle /cards/named?exact=...
		if name := r.URL.Query().Get("exact"); name != "" {
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
			return
		}

		// Handle /cards/:set/:number
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/cards/"), "/")
		if len(parts) == 2 {
			key := parts[0] + "/" + parts[1]
			c, ok := cards[key]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{
					"object":  "error",
					"details": fmt.Sprintf("Card not found: %s", key),
				})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(c)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestFetchInstant(t *testing.T) {
	server := newNameServer(t, map[string]scryfallResponse{
		"Lightning Bolt": {
			Object: "card", Name: "Lightning Bolt", ManaCost: "{R}",
			TypeLine: "Instant", OracleText: "Lightning Bolt deals 3 damage to any target.",
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	c, err := client.FetchCard(byName("Lightning Bolt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Name != "Lightning Bolt" {
		t.Errorf("name: got %q, want %q", c.Name, "Lightning Bolt")
	}
	if c.ManaCost.String() != "R" {
		t.Errorf("mana cost: got %q, want %q", c.ManaCost, "R")
	}
	if c.Stats != nil {
		t.Errorf("stats: expected nil for instant")
	}
	if c.Loyalty != nil {
		t.Errorf("loyalty: expected nil for instant")
	}
}

func TestFetchCreature(t *testing.T) {
	server := newNameServer(t, map[string]scryfallResponse{
		"Tarmogoyf": {
			Object: "card", Name: "Tarmogoyf", ManaCost: "{1}{G}",
			TypeLine: "Creature — Lhurgoyf",
			OracleText: "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.",
			Power: ptr("*"), Toughness: ptr("1+*"),
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	c, err := client.FetchCard(byName("Tarmogoyf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Stats == nil {
		t.Fatal("stats: expected non-nil for creature")
	}
	if c.Stats.Power != "*" || c.Stats.Toughness != "1+*" {
		t.Errorf("stats: got %v", *c.Stats)
	}
}

func TestFetchPlaneswalker(t *testing.T) {
	server := newNameServer(t, map[string]scryfallResponse{
		"Jace, the Mind Sculptor": {
			Object: "card", Name: "Jace, the Mind Sculptor", ManaCost: "{2}{U}{U}",
			TypeLine:   "Legendary Planeswalker — Jace",
			OracleText: "+2: Look at the top card.\n−12: Exile all cards.",
			Loyalty:    ptr("3"),
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	c, err := client.FetchCard(byName("Jace, the Mind Sculptor"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Loyalty == nil || *c.Loyalty != "3" {
		t.Errorf("loyalty: got %v", c.Loyalty)
	}
	if c.ManaCost.String() != "2UU" {
		t.Errorf("mana cost: got %q", c.ManaCost)
	}
}

func TestFetchNotFound(t *testing.T) {
	server := newNameServer(t, map[string]scryfallResponse{})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	_, err := client.FetchCard(byName("Nonexistent Card"))
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
	client.FetchCard(byName("Lightning Bolt"))

	if receivedPath != "/cards/named?exact=Lightning+Bolt" {
		t.Errorf("request path: got %q", receivedPath)
	}
	if ua := receivedHeaders.Get("User-Agent"); ua != "mtg-proxy/1.0" {
		t.Errorf("User-Agent: got %q", ua)
	}
	if accept := receivedHeaders.Get("Accept"); accept != "application/json" {
		t.Errorf("Accept: got %q", accept)
	}
}

func TestFetchBySetAndCollector(t *testing.T) {
	server := newNameServer(t, map[string]scryfallResponse{
		"6ed/273": {
			Object: "card", Name: "Ankh of Mishra", ManaCost: "{2}",
			TypeLine: "Artifact", OracleText: "Whenever a land enters, Ankh of Mishra deals 2 damage to that land's controller.",
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	c, err := client.FetchCard(bySet("Ankh of Mishra", "6ed", "273"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Name != "Ankh of Mishra" {
		t.Errorf("name: got %q", c.Name)
	}
}

func TestFetchBySetSendsCorrectPath(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewEncoder(w).Encode(scryfallResponse{
			Object: "card", Name: "Ankh of Mishra", ManaCost: "{2}", TypeLine: "Artifact",
		})
	}))
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	client.FetchCard(bySet("Ankh of Mishra", "6ed", "273"))

	if receivedPath != "/cards/6ed/273" {
		t.Errorf("request path: got %q, want %q", receivedPath, "/cards/6ed/273")
	}
}

func TestFetchFallsBackToNameWhenNoSet(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.RequestURI()
		json.NewEncoder(w).Encode(scryfallResponse{
			Object: "card", Name: "Lightning Bolt", ManaCost: "{R}", TypeLine: "Instant",
		})
	}))
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	client.FetchCard(byName("Lightning Bolt"))

	if !strings.Contains(receivedPath, "/cards/named?exact=") {
		t.Errorf("expected name lookup fallback, got path: %q", receivedPath)
	}
}

func TestFetchLand(t *testing.T) {
	server := newNameServer(t, map[string]scryfallResponse{
		"Island": {
			Object: "card", Name: "Island", ManaCost: "",
			TypeLine: "Basic Land — Island", OracleText: "({T}: Add {U}.)",
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	c, err := client.FetchCard(byName("Island"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.ManaCost.IsEmpty() {
		t.Errorf("mana cost: expected empty for land, got %q", c.ManaCost)
	}
}

func TestFetchThrottlesRequests(t *testing.T) {
	server := newNameServer(t, map[string]scryfallResponse{
		"Lightning Bolt": {
			Object: "card", Name: "Lightning Bolt", ManaCost: "{R}", TypeLine: "Instant",
		},
	})
	defer server.Close()

	client := scryfall.NewClient(server.URL)
	entry := byName("Lightning Bolt")

	start := time.Now()
	client.FetchCard(entry)
	firstDuration := time.Since(start)

	start = time.Now()
	client.FetchCard(entry)
	secondDuration := time.Since(start)

	if secondDuration < 80*time.Millisecond {
		t.Errorf("second request was too fast (%v), expected >= 80ms throttle", secondDuration)
	}
	if firstDuration > 50*time.Millisecond {
		t.Errorf("first request was unexpectedly slow (%v)", firstDuration)
	}
}
