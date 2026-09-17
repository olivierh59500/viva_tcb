package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	vivatcb "viva_tcb"
)

func main() {
	ebiten.SetWindowSize(vivatcb.ScreenWidth, vivatcb.ScreenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("VIVA TCB! - Go/Ebitengine Port")
	ebiten.SetScreenClearedEveryFrame(false)

	if err := ebiten.RunGame(newDrawOnUpdateGame(vivatcb.NewGame())); err != nil {
		log.Fatal(err)
	}
}
