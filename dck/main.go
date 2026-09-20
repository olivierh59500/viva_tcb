// Package vivatcb implements the VIVA TCB Ebitengine demo.
package vivatcb

import originalassets "viva_tcb"

import (
	"bytes"

	"fmt"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/olivierh59500/ym-player/pkg/stsound"
)

const (
	// ScreenWidth and ScreenHeight are the demo's fixed logical dimensions.
	ScreenWidth  = 768
	ScreenHeight = 540

	fontCharWidth  = 42
	fontCharHeight = 40
	sampleRate     = 48000

	// These phases preserve the texture origin of the former 16-screen-wide
	// pre-rendered tile canvas without allocating that roughly 400 MiB image.
	tilePhaseX = float64(ScreenWidth * 8)
	tilePhaseY = float64(ScreenHeight * 8)
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

var repeatingQuadIndices = [...]uint16{0, 1, 2, 1, 2, 3}

// Game contains the complete state of the demo.
type Game struct {
	initialized bool

	logoImg        *ebiten.Image
	titleImg       *ebiten.Image
	rasterImg      *ebiten.Image
	tileImg        *ebiten.Image
	fontImg        *ebiten.Image
	scrollPrograms [4]*scrolling.Scrolling

	titleCanvas *ebiten.Image
	topBar      *ebiten.Image

	audioContext *audio.Context
	audioPlayer  *audio.Player
	ymPlayer     *YMPlayer
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

	fxFlag int
	posXi  float64
	posZi  float64
	posRi  float64

	initX float64
	initR float64

	text1 string
	text2 string
	text3 string
	text4 string
}

// YMPlayer adapts ym-player's mono int16 stream to Ebitengine's interleaved
// little-endian stereo PCM reader.
type YMPlayer struct {
	player *stsound.StSound
	buffer []int16
	mutex  sync.Mutex
	loop   bool
}

// NewYMPlayer creates an allocation-free YM audio reader after setup.
func NewYMPlayer(data []byte, rate int, loop bool) (*YMPlayer, error) {
	player := stsound.CreateWithRate(rate)
	if err := player.LoadMemory(data); err != nil {
		player.Destroy()
		return nil, fmt.Errorf("load YM data: %w", err)
	}
	player.SetLoopMode(loop)

	return &YMPlayer{
		player: player,
		buffer: make([]int16, 4096),
		loop:   loop,
	}, nil
}

// Read writes PCM16 stereo frames directly into p.
func (y *YMPlayer) Read(p []byte) (n int, err error) {
	y.mutex.Lock()
	defer y.mutex.Unlock()

	samplesNeeded := len(p) / 4
	if samplesNeeded == 0 {
		return 0, nil
	}
	if y.player == nil {
		clear(p[:samplesNeeded*4])
		return samplesNeeded * 4, io.EOF
	}

	processed := 0
	for processed < samplesNeeded {
		chunkSize := min(samplesNeeded-processed, len(y.buffer))
		if !y.player.Compute(y.buffer[:chunkSize], chunkSize) && !y.loop {
			clear(p[processed*4 : samplesNeeded*4])
			err = io.EOF
			break
		}

		for i := 0; i < chunkSize; i++ {
			sample := y.buffer[i] / 2
			offset := (processed + i) * 4
			p[offset] = byte(sample)
			p[offset+1] = byte(sample >> 8)
			p[offset+2] = byte(sample)
			p[offset+3] = byte(sample >> 8)
		}
		processed += chunkSize
	}

	return samplesNeeded * 4, err
}

// Close releases the synthesizer. It is safe to call more than once.
func (y *YMPlayer) Close() error {
	y.mutex.Lock()
	defer y.mutex.Unlock()

	if y.player != nil {
		y.player.Destroy()
		y.player = nil
	}
	return nil
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

	ym, err := NewYMPlayer(ymData, sampleRate, true)
	if err != nil {
		log.Printf("cannot create YM player: %v", err)
		return
	}
	g.ymPlayer = ym

	player, err := g.audioContext.NewPlayer(ym)
	if err != nil {
		log.Printf("cannot create audio player: %v", err)
		if closeErr := ym.Close(); closeErr != nil {
			log.Printf("cannot close YM player: %v", closeErr)
		}
		g.ymPlayer = nil
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

func mapCharToFont(charCode int) int {
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

	if g.fxFlag >= 1 {
		g.posXi += 0.008
	}
	if g.fxFlag >= 2 {
		g.posZi += 0.003
	}
	if g.fxFlag >= 3 {
		g.posRi += 0.005
	}
	if g.posXi >= 2 {
		g.fxFlag = 2
	}
	if g.posZi >= 1.5 {
		g.fxFlag = 3
	}

	if g.fxFlag == 0 {
		g.initX -= 4
		if g.initX <= -float64(ScreenWidth*2) {
			g.initR--
		}
		if g.initR <= -45 {
			g.fxFlag = 1
		}
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

	if g.fxFlag >= 1 {
		zoom := 0.5 + math.Abs(math.Sin(g.posZi)*2.5)
		rotation := (90 * math.Cos(g.posRi*4-math.Cos(g.posRi-0.01))) * 0.3 * math.Pi / 180
		posXCurve := math.Cos(g.posXi - 0.1)
		oscX := (float64(ScreenWidth) / 4) * math.Cos(g.posXi*4-posXCurve)
		oscY := (float64(ScreenHeight) / 2.7) * -math.Sin(g.posXi*2.3-posXCurve)
		drawRepeatingRotozoom(
			screen,
			g.tileImg,
			float64(ScreenWidth)/2+oscX,
			float64(ScreenHeight)/2+oscY,
			zoom,
			rotation,
			tilePhaseX,
			tilePhaseY,
		)
	} else {
		drawRepeatingRotozoom(
			screen,
			g.tileImg,
			g.initX+float64(ScreenWidth)/2,
			float64(ScreenHeight)/2,
			1,
			g.initR*0.3*math.Pi/180,
			tilePhaseX,
			tilePhaseY,
		)
	}

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

func drawRepeatingRotozoom(dst, texture *ebiten.Image, centerX, centerY, zoom, rotation, phaseX, phaseY float64) {
	composite.Repeat(dst, texture, composite.Repetition{CenterX: centerX, CenterY: centerY, Zoom: zoom, Rotation: rotation, PhaseX: phaseX, PhaseY: phaseY})
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
