package profanity

import (
	"regexp"
	"strings"
	"testing"
)

func TestHasBlockedWords(t *testing.T) {
	word := "pluto"
	pattern := `(?i)\b` + regexp.QuoteMeta(strings.TrimSpace(word)) + `\b`
	blockedRegexes = append(blockedRegexes, regexp.MustCompile(pattern))

	tests := []struct {
		input    string
		expected bool
	}{
		{"H3ll0 pluto W0rld!", true},
		{"Th1s 1s 4 t3st", false},
		{"iL0v3Plut0", true},
		{"I_love_pluto_dwarf_planet", true},
		{"whoLetPlutoOut", true},
		{"who let plUto out", true},
		{"pluto", true},
		{"this-pluto-dwarf-planet", true},
		{"thisPluto_dwarf_planet", true},
		{"this_plutoDwarfPlanet", true},
		{"thisPluto-dwarf-planet", true},
		{"this-plutoDwarfPlanet", true},
		{"there are other planets", false},
	}

	for _, test := range tests {
		result := hasBlockedWords(test.input, blockedRegexes)
		if result != test.expected {
			t.Errorf("hasBlockedWords(%q) = %t, expected %t", test.input, result, test.expected)
		}
	}
}
