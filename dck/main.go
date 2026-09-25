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
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/sound"
	"github.com/olivierh59500/democonstructionkit/sprites"

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

// Game contains the complete state of the demo.
type Game struct {
	initialized bool

	logoImg       *ebiten.Image
	titleImg      *ebiten.Image
	rasterImg     *ebiten.Image
	tileImg       *ebiten.Image
	backdrop      *composite.RotozoomBackground
	fontImg       *ebiten.Image
	pseudoScroll  *scrolling.Scrolling
	logoFormation *sprites.RecurrentFormation

	titleCanvas *ebiten.Image
	topBar      *ebiten.Image

	audioContext *audio.Context
	audioPlayer  *audio.Player
	musicStream  *sound.Stream
	audioReady   bool
	musicStarted bool

	logoX        float64
	hold         int
	rasterMotion *motion.WrapBank

	loopCounter int

	text1 string
	text2 string
	text3 string
	text4 string
}

// NewGame creates a demo whose graphics and audio are initialized lazily from
// the game loop. Delaying this work is required while Android loads libgojni.
func NewGame() *Game {
	return &Game{
		logoX: 1.5,
		hold:  970,
		text1: "                BILIZIR FROM DMA PRESENTS HIS LATEST GOLANG/EBITEN CONVERSION. THE ORIGINAL IDEA AND SCREEN IS FROM MELLOW MAN. THIS IS ANOTHER TRIBUTE                              TO THE ST LEGENDS 'TCB' !       ",
		text2: "                               WITH EBITEN IT'S TOO EASY TO CREATE THIS KIND OF OLDSCHOOL DEMO, THIS IS ANOTHER NATIVE SCREEN BUILT WITH GOLANG AND EBITEN.                     WELL, IT'S NOT EXACTLY THE SAME EFFECTS, BUT IT'S CLOSE.             ",
		text3: "                                            YES... THERE IS A THIRD SCROLLTEXT IN THIS SCREEN..... MUCH LIKE ON THE ORIGINAL TCB FULLSCREEN DEMO, WE HAVE MULTIPLE DIFFERENT SCROLLERS... AND THESE ALL HAVE THIS COOL EFFECT ON THEM.... I HOPE YOU LIKE IT.....            ",
		text4: "                                                                               I GUESS WE SHOULD HAVE SOME GREETINGS, AS IT IS A DEMOSCREEN IN THE OLD SCHOOL STYLEE! SO HERE THEY ARE.... THE GREETZ GO OUT TO:  ALL MEMBERS OF DMA (PDM, COCO, JINX, CORWIN, DWORKIN) - MELLOW MAN - NONAMENO - THE UNION (MAD MAX FROM TEX FOR THE MUSIC) - COMMODOREBLOG - ELKMOOSE AND ANYONE ELSE I MAY HAVE MISSED!    LETZ WRAP..............       ",
	}
}

// NewSilentGame retains the visual program while skipping device audio during
// deterministic native-frame capture.
func NewSilentGame() *Game {
	game := NewGame()
	game.audioReady = true
	return game
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
	g.rasterMotion, err = motion.NewWrapBank(presets.VivaRasterWrapConfig())
	if err != nil {
		return err
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
	pseudo := presets.VivaPseudo3D(scrolling.Face{Atlas: g.fontImg, Metrics: metrics}, nil,
		[4]string{g.text1, g.text2, g.text3, g.text4}, ScreenWidth, ScreenHeight)
	g.pseudoScroll, err = scrolling.New(scrolling.Config{Pseudo3D: &pseudo})
	if err != nil {
		return err
	}
	formation := presets.VivaLogoFormation(g.logoImg, ScreenWidth, ScreenHeight, float64(ScreenHeight/2-64)/4)
	g.logoFormation, err = sprites.NewRecurrentFormation(formation)
	if err != nil {
		return err
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

	g.rasterMotion.Step()

	g.loopCounter++
	if err := g.logoFormation.Update(float64(g.loopCounter)); err != nil {
		return err
	}
	return g.pseudoScroll.Update(kit.Frame{Tick: uint64(g.loopCounter), Time: float64(g.loopCounter) / 60})
}

// Draw renders the current immutable update state.
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.initialized {
		screen.Fill(color.Black)
		return
	}

	g.backdrop.Draw(screen)

	g.pseudoScroll.Draw(screen)

	var topBarOp ebiten.DrawImageOptions
	screen.DrawImage(g.topBar, &topBarOp)
	g.drawTitle(screen)
	g.logoFormation.Draw(screen)
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	g.titleCanvas.Fill(color.Black)
	for _, rasterY := range [...]float64{g.rasterMotion.At(0), g.rasterMotion.At(1), g.rasterMotion.At(1) + 72} {
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

// Layout returns the demo's fixed logical size. Ebitengine letterboxes it on
// wider displays, preserving the original pixels and aspect ratio.
func (g *Game) Layout(_, _ int) (int, int) {
	return ScreenWidth, ScreenHeight
}
