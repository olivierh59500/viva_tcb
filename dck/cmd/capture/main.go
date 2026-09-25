// Command capture records deterministic native frames of the DCK Viva scene.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	vivatcb "viva_tcb/dck"
)

func main() {
	framesFlag := flag.String("frames", "1,60,240,600", "comma-separated capture ticks")
	out := flag.String("out", "captures/native", "capture directory")
	flag.Parse()
	var frames []int
	for _, part := range strings.Split(*framesFlag, ",") {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 || len(frames) > 0 && n <= frames[len(frames)-1] {
			fmt.Fprintln(os.Stderr, "frames must be increasing nonnegative integers")
			os.Exit(1)
		}
		frames = append(frames, n)
	}
	if err := capture.Run(capture.Config{Directory: *out, Frames: frames, Width: vivatcb.ScreenWidth, Height: vivatcb.ScreenHeight}, func() (ebiten.Game, error) {
		return vivatcb.NewSilentGame(), nil
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
