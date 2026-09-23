package vivatcb

import "testing"

func legacymapCharToFont(charCode int) int {
	if charCode >= 'a' && charCode <= 'z' {
		charCode -= 'a' - 'A'
	}
	switch {
	case charCode == ' ':
		return 0
	case charCode >= '!' && charCode <= '@':
		return charCode - ' '
	case charCode >= 'A' && charCode <= 'Z':
		return charCode - 'A' + 33
	default:
		return 0
	}
}
func TestSharedmapCharToFontMatchesOriginal(t *testing.T) {
	for r := 0; r < 256; r++ {
		if got, want := mapCharToFont(int(r)), legacymapCharToFont(int(r)); got != want {
			t.Fatalf("rune %U: got %d, want %d", r, got, want)
		}
	}
}
