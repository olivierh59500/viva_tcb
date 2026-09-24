// Package vivatcb implements the VIVA TCB Ebitengine demo.
package vivatcb

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	originalassets "viva_tcb"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/sound"

	_ "image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	audio "github.com/olivierh59500/democonstructionkit/sound/output"
)

const (
	// ScreenWidth and ScreenHeight are the demo's fixed logical dimensions.
	ScreenWidth  = 768
	ScreenHeight = 540

	fontCharWidth  = 42
	fontCharHeight = 40
	sampleRate     = 48000
)

var assets = originalassets.
	DCKAssetAssets()

var ymData = originalassets.DCKAssetYmData()

var (
	scrollerZSinStep, scrollerZCosStep           = math.Sincos(0.75)
	scrollerXSinStep, scrollerXCosStep           = math.Sincos(18)
	scrollerYSinStep, scrollerYCosStep           = math.Sincos(0.7)
	logoXSinStep, logoXCosStep                   = math.Sincos(0.2)
	logoXSecondarySinStep, logoXSecondaryCosStep = math.Sincos(1.0 / 60.0)
	logoYSinStep, logoYCosStep                   = math.Sincos(5.0 / 37.0)
	logoYSecondarySinStep, logoYSecondaryCosStep = math.Sincos(5.0 / 17.0)
)

// Game contains the complete state of the demo.
type Game struct {
	initialized bool

	logoImg        *ebiten.Image
	titleImg       *ebiten.Image
	rasterImg      *ebiten.Image
	tileImg        *ebiten.Image
	backdrop       *composite.RotozoomBackground
	fontImg        *ebiten.Image
	scrollPrograms [4]*scrolling.Scrolling

	titleCanvas *ebiten.Image
	topBar      *ebiten.Image

	audioContext *audio.Context
	audioPlayer  *audio.Player
	musicStream  *sound.Stream
	audioReady   bool
	musicStarted bool

	logoX    float64
	hold     int
	rasterY1 float64
	rasterY2 float64

	loopCounter int

	scrollX1 float64
	scrollX2 float64
	scrollX3 float64
	scrollX4 float64

	text1 string
	text2 string
	text3 string
	text4 string
}

// NewGame creates a demo whose graphics and audio are initialized lazily from
// the game loop. Delaying this work is required while Android loads libgojni.
func NewGame() *Game {
	return &Game{
		logoX:    1.5,
		hold:     970,
		rasterY1: 0,
		rasterY2: 72,
		text1:    "                BILIZIR FROM DMA PRESENTS HIS LATEST GOLANG/EBITEN CONVERSION. THE ORIGINAL IDEA AND SCREEN IS FROM MELLOW MAN. THIS IS ANOTHER TRIBUTE                              TO THE ST LEGENDS 'TCB' !       ",
		text2:    "                               WITH EBITEN IT'S TOO EASY TO CREATE THIS KIND OF OLDSCHOOL DEMO, THIS IS ANOTHER NATIVE SCREEN BUILT WITH GOLANG AND EBITEN.                     WELL, IT'S NOT EXACTLY THE SAME EFFECTS, BUT IT'S CLOSE.             ",
		text3:    "                                            YES... THERE IS A THIRD SCROLLTEXT IN THIS SCREEN..... MUCH LIKE ON THE ORIGINAL TCB FULLSCREEN DEMO, WE HAVE MULTIPLE DIFFERENT SCROLLERS... AND THESE ALL HAVE THIS COOL EFFECT ON THEM.... I HOPE YOU LIKE IT.....            ",
		text4:    "                                                                               I GUESS WE SHOULD HAVE SOME GREETINGS, AS IT IS A DEMOSCREEN IN THE OLD SCHOOL STYLEE! SO HERE THEY ARE.... THE GREETZ GO OUT TO:  ALL MEMBERS OF DMA (PDM, COCO, JINX, CORWIN, DWORKIN) - MELLOW MAN - NONAMENO - THE UNION (MAD MAX FROM TEX FOR THE MUSIC) - COMMODOREBLOG - ELKMOOSE AND ANYONE ELSE I MAY HAVE MISSED!    LETZ WRAP..............       ",
	}
}

func loadImage(path string) (*ebiten.Image, error) {
	data, err := assets.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(decoded), nil
}

// Init initializes the graphics resources once. Audio is deliberately kept
// separate and is opened only from the first Update call.
func (g *Game) Init() error {
	if g.initialized {
		return nil
	}

	var err error
	if g.logoImg, err = loadImage("assets/logo.png"); err != nil {
		return fmt.Errorf("load logo: %w", err)
	}
	if g.titleImg, err = loadImage("assets/title.png"); err != nil {
		return fmt.Errorf("load title: %w", err)
	}
	if g.rasterImg, err = loadImage("assets/raster.png"); err != nil {
		return fmt.Errorf("load raster: %w", err)
	}
	if g.tileImg, err = loadImage("assets/tcb_tile.png"); err != nil {
		return fmt.Errorf("load tile: %w", err)
	}
	program, err := presets.NewVivaRotozoom(presets.DefaultVivaRotozoomConfig(ScreenWidth, ScreenHeight))
	if err != nil {
		return err
	}
	g.backdrop, err = composite.NewRotozoomBackground(composite.RotozoomBackgroundConfig{Image: g.tileImg, Program: program})
	if err != nil {
		return err
	}
	if g.fontImg, err = loadImage("assets/font.png"); err != nil {
		return fmt.Errorf("load font: %w", err)
	}

	spec, _ := presets.FindFont("viva_tcb")
	metrics, err := spec.Build(g.fontImg.Bounds())
	if err != nil {
		return err
	}
	for i, text := range [...]string{g.text1, g.text2, g.text3, g.text4} {
		g.scrollPrograms[i], err = scrolling.New(scrolling.Config{Text: text, Advance: 64, Fonts: map[string]scrolling.Face{"default": {Atlas: g.fontImg, Metrics: metrics}}})
		if err != nil {
			return err
		}
	}

	g.titleCanvas = ebiten.NewImage(g.titleImg.Bounds().Dx(), g.titleImg.Bounds().Dy())
	g.topBar = ebiten.NewImage(ScreenWidth, 64)
	g.topBar.Fill(color.Black)
	g.initialized = true
	return nil
}

func (g *Game) initAudio() {
	g.audioContext = audio.NewContext(sampleRate)

	music, err := sound.Open("music.ym", ymData, sound.Options{SampleRate: sampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 0.5})
	if err != nil {
		log.Printf("cannot open music: %v", err)
		return
	}
	g.musicStream = music

	player, err := g.audioContext.NewPlayer(music)
	if err != nil {
		log.Printf("cannot create audio player: %v", err)
		if closeErr := music.Close(); closeErr != nil {
			log.Printf("cannot close music stream: %v", closeErr)
		}
		g.musicStream = nil
		return
	}
	g.audioPlayer = player
}

func (g *Game) startMusic() {
	if g.musicStarted || g.audioPlayer == nil {
		return
	}
	g.audioPlayer.Play()
	g.musicStarted = true
}

var mapCharToFont = func() func(int) int {
	lookup, err := presets.TileLookup("viva_tcb", false)
	if err != nil {
		panic(err)
	}
	return func(ch int) int { index, _ := lookup(rune(ch)); return index }
}()

func stepSinCosBackward(sinValue, cosValue, sinStep, cosStep float64) (float64, float64) {
	return sinValue*cosStep - cosValue*sinStep, cosValue*cosStep + sinValue*sinStep
}

func stepSinCosForward(sinValue, cosValue, sinStep, cosStep float64) (float64, float64) {
	return sinValue*cosStep + cosValue*sinStep, cosValue*cosStep - sinValue*sinStep
}

func (g *Game) drawScroller(dst *ebiten.Image, text string, scrollX float64, scrollerID int, baseY, t, horizontalWave, verticalWave float64) {
	if len(text) == 0 {
		return
	}
	firstIndex := int(scrollX / 64)
	maxIndex := firstIndex + 8
	zSin, zCos := math.Sincos((t + float64(maxIndex)*.15) * 5)
	xSin, xCos := math.Sincos(t*7 + float64(maxIndex)*18)
	ySin, yCos := math.Sincos((t + float64(maxIndex)*.1) * 7)
	advance := func() {
		zSin, zCos = stepSinCosBackward(zSin, zCos, scrollerZSinStep, scrollerZCosStep)
		xSin, xCos = stepSinCosBackward(xSin, xCos, scrollerXSinStep, scrollerXCosStep)
		ySin, yCos = stepSinCosBackward(ySin, yCos, scrollerYSinStep, scrollerYCosStep)
	}
	for i := maxIndex; i >= len(text); i-- {
		advance()
	}
	state := scrolling.IdentityState()
	state.First = firstIndex
	state.End = maxIndex + 1
	state.Reverse = true
	state.Time = t
	state.Position = scrollX
	state.Map = func(sample scrolling.Sample, op *ebiten.DrawImageOptions) bool {
		currentZSin, currentXSin, currentYSin := zSin, xSin, ySin
		advance()
		z := currentZSin*.5 + 1.5
		drawX := math.Floor((float64(sample.Index)*64 - 40 - currentXSin*32*horizontalWave - scrollX) * 2)
		drawY := math.Floor(currentYSin*42*verticalWave + baseY - z*32)
		scale := z
		if scrollerID == 1 || scrollerID == 2 {
			scale = 3 - z
		}
		if drawX < -100 || drawX > ScreenWidth+100 || drawY < -100 || drawY > ScreenHeight+100 || scale <= .1 {
			return false
		}
		op.GeoM.Reset()
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(drawX, drawY)
		op.ColorScale.Scale(1, 1, 1, .9)
		return true
	}
	g.scrollPrograms[scrollerID-1].DrawAt(dst, state)
}

func advanceScroller(scrollX float64, text string) float64 {
	if len(text) == 0 {
		return 0
	}
	scrollX += 4
	if limit := float64(len(text) * 64); scrollX >= limit {
		scrollX -= limit
	}
	return scrollX
}

// Update advances the demo by one fixed 60 Hz tick.
func (g *Game) Update() error {
	if !g.initialized {
		if err := g.Init(); err != nil {
			return err
		}
	}
	if !g.audioReady {
		g.audioReady = true
		g.initAudio()
		g.startMusic()
	}

	if err := g.backdrop.Update(kit.Frame{Tick: uint64(g.loopCounter + 1), Time: float64(g.loopCounter+1) / 60, Delta: 1.0 / 60}); err != nil {
		return err
	}

	if g.hold > 0 {
		g.hold--
	} else {
		g.logoX += 0.0125
	}

	g.rasterY1 -= 2
	g.rasterY2 -= 2
	if g.rasterY1 <= -72 {
		g.rasterY1 = 72
	}
	if g.rasterY2 <= -72 {
		g.rasterY2 = 72
	}

	g.loopCounter++
	g.scrollX1 = advanceScroller(g.scrollX1, g.text1)
	g.scrollX2 = advanceScroller(g.scrollX2, g.text2)
	g.scrollX3 = advanceScroller(g.scrollX3, g.text3)
	g.scrollX4 = advanceScroller(g.scrollX4, g.text4)
	return nil
}

// Draw renders the current immutable update state.
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.initialized {
		screen.Fill(color.Black)
		return
	}

	g.backdrop.Draw(screen)

	t := float64(g.loopCounter)/60 + 19
	wave := math.Sin(t*0.25)*0.5 + 0.5
	horizontalWave := math.Sqrt(1 - wave*wave)
	verticalWave := math.Sin(t*0.5)*0.5 + 0.5
	g.drawScroller(screen, g.text1, g.scrollX1, 1, 500, t, horizontalWave, verticalWave)
	g.drawScroller(screen, g.text2, g.scrollX2, 2, 250, t, horizontalWave, verticalWave)
	g.drawScroller(screen, g.text3, g.scrollX3, 3, 375, t, horizontalWave, verticalWave)
	g.drawScroller(screen, g.text4, g.scrollX4, 4, 125, t, horizontalWave, verticalWave)

	var topBarOp ebiten.DrawImageOptions
	screen.DrawImage(g.topBar, &topBarOp)
	g.drawTitle(screen)
	g.drawLogos(screen)
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	g.titleCanvas.Fill(color.Black)
	for _, rasterY := range [...]float64{g.rasterY1, g.rasterY2, g.rasterY2 + 72} {
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(24, 1)
		op.GeoM.Translate(0, rasterY)
		composite.Instance{Image: g.rasterImg, Options: op}.Draw(g.titleCanvas)
	}
	var titleOp ebiten.DrawImageOptions
	composite.Instance{Image: g.titleImg, Options: titleOp}.Draw(g.titleCanvas)

	var canvasOp ebiten.DrawImageOptions
	canvasOp.GeoM.Translate(64+ScreenWidth*math.Cos(g.logoX), 14)
	composite.Instance{Image: g.titleCanvas, Options: canvasOp}.Draw(screen)
}

func (g *Game) drawLogos(screen *ebiten.Image) {
	midX := float64(ScreenWidth/2-32) / 2
	midY := 24 + float64(ScreenHeight/2-32)/2
	incY := float64(ScreenHeight/2-64) / 4
	base := float64(g.loopCounter)
	xSin, xCos := math.Sincos(base / 25)
	xSecondarySin, xSecondaryCos := math.Sincos(base / 300)
	ySin, yCos := math.Sincos(base / 37)
	ySecondarySin, ySecondaryCos := math.Sincos(base / 17)

	for range 10 {
		spX := midX + midX*xSin*xSecondaryCos
		spY := midY + incY*ySin + incY*ySecondaryCos

		var op ebiten.DrawImageOptions
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(spX*2, spY*2)
		composite.Instance{Image: g.logoImg, Options: op}.Draw(screen)

		xSin, xCos = stepSinCosForward(xSin, xCos, logoXSinStep, logoXCosStep)
		xSecondarySin, xSecondaryCos = stepSinCosForward(
			xSecondarySin,
			xSecondaryCos,
			logoXSecondarySinStep,
			logoXSecondaryCosStep,
		)
		ySin, yCos = stepSinCosForward(ySin, yCos, logoYSinStep, logoYCosStep)
		ySecondarySin, ySecondaryCos = stepSinCosForward(
			ySecondarySin,
			ySecondaryCos,
			logoYSecondarySinStep,
			logoYSecondaryCosStep,
		)
	}
}

// Layout returns the demo's fixed logical size. Ebitengine letterboxes it on
// wider displays, preserving the original pixels and aspect ratio.
func (g *Game) Layout(_, _ int) (int, int) {
	return ScreenWidth, ScreenHeight
}
