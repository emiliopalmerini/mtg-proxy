package card_test

import (
	"testing"

	"github.com/epalmerini/mtg-proxy/internal/card"
)

func TestIsBasicLand(t *testing.T) {
	tests := []struct {
		typeLine card.TypeLine
		want     bool
	}{
		{"Basic Land — Plains", true},
		{"Basic Land — Island", true},
		{"Basic Land — Swamp", true},
		{"Basic Land — Mountain", true},
		{"Basic Land — Forest", true},
		{"Basic Snow Land — Island", true},
		{"Land — Island", false},
		{"Creature — Human Wizard", false},
		{"Instant", false},
		{"Legendary Planeswalker — Jace", false},
		{"Artifact", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.typeLine), func(t *testing.T) {
			if got := tt.typeLine.IsBasicLand(); got != tt.want {
				t.Errorf("TypeLine(%q).IsBasicLand() = %v, want %v", tt.typeLine, got, tt.want)
			}
		})
	}
}
