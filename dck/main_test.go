package vivatcb

import (
	"encoding/binary"
	"testing"

	"github.com/olivierh59500/democonstructionkit/sound"
)

func TestMusicStreamReadProducesStereoWithoutAllocating(t *testing.T) {
	player, err := sound.Open("music.ym", ymData, sound.Options{SampleRate: sampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := player.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	buffer := make([]byte, 4096*4)
	read := func() {
		n, err := player.Read(buffer)
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if n != len(buffer) {
			t.Fatalf("Read bytes = %d, want %d", n, len(buffer))
		}
	}

	read()
	for i := 0; i < len(buffer); i += 4 {
		left := binary.LittleEndian.Uint16(buffer[i : i+2])
		right := binary.LittleEndian.Uint16(buffer[i+2 : i+4])
		if left != right {
			t.Fatalf("frame %d is not mono duplicated to stereo: %d != %d", i/4, left, right)
		}
	}

	if allocations := testing.AllocsPerRun(20, read); allocations != 0 {
		t.Fatalf("Read allocations = %v, want 0", allocations)
	}
}

func TestMapCharToFont(t *testing.T) {
	tests := map[byte]int{
		' ': 0,
		'!': 1,
		'@': 32,
		'A': 33,
		'Z': 58,
		'a': 33,
		'z': 58,
		'^': 0,
	}
	for char, want := range tests {
		if got := mapCharToFont(int(char)); got != want {
			t.Errorf("mapCharToFont(%q) = %d, want %d", char, got, want)
		}
	}
}

func TestNewGameDefersPlatformResources(t *testing.T) {
	game := NewGame()
	if game.initialized || game.audioReady || game.audioContext != nil || game.audioPlayer != nil {
		t.Fatal("NewGame initialized graphics or audio before the game loop")
	}
}
