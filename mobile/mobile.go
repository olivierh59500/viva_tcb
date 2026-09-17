//go:build android || ios

// Package mobile exposes VIVA TCB to Ebitengine's native mobile view.
package mobile

import (
	"github.com/hajimehoshi/ebiten/v2"
	enginemobile "github.com/hajimehoshi/ebiten/v2/mobile"
	vivatcb "viva_tcb"
)

func init() {
	ebiten.SetScreenClearedEveryFrame(false)
	enginemobile.SetGame(vivatcb.NewGame())
}

// Dummy ensures gomobile emits bindings for this package.
func Dummy() {}
