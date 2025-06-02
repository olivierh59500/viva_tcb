package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth    = 768
	screenHeight   = 540
	tileWidth      = 256
	tileHeight     = 256
	canvasWidth    = screenWidth * 16
	canvasHeight   = screenHeight * 16
	fontCharWidth  = 42
	fontCharHeight = 40
)

//go:embed assets/*
var assets embed.FS

// Game représente l'état principal du jeu
type Game struct {
	// Images
	logoImg   *ebiten.Image // Remplace spriteImg
	titleImg  *ebiten.Image
	rasterImg *ebiten.Image
	tileImg   *ebiten.Image
	fontImg   *ebiten.Image

	// Surfaces/Canvas virtuels
	tileCanvas   *ebiten.Image
	spriteCanvas *ebiten.Image
	titleCanvas  *ebiten.Image

	// Audio
	audioContext *audio.Context
	audioPlayer  *audio.Player

	// Variables d'animation
	logoX    float64
	hold     int
	rasterY1 float64
	rasterY2 float64

	// Timing
	startTime   time.Time
	loopCounter int

	// Positions des scrollers
	scrollX1 float64
	scrollX2 float64
	scrollX3 float64
	scrollX4 float64

	// Variables d'effet
	fxFlag int
	posXi  float64
	posZi  float64
	posRi  float64

	// Animation d'intro
	initX float64
	initR float64

	// Textes
	text1 string
	text2 string
	text3 string
	text4 string

	// Debug
	debugFrameCount int
}

// NewGame crée une nouvelle instance du jeu
func NewGame() *Game {
	g := &Game{
		logoX:     1.5,
		hold:      970,
		rasterY1:  0,
		rasterY2:  72,
		initX:     0,
		initR:     0,
		fxFlag:    0,
		startTime: time.Now(),
	}

	// Initialiser les textes
	g.text1 = "                BILIZIR FROM DMA PRESENTS HIS LATEST GOLANG/EBITEN CONVERSION. THE ORIGINAL IDEA AND SCREEN IS FROM MELLOW MAN. THIS IS ANOTHER TRIBUTE                              TO THE ST LEGENDS 'TCB' !       "
	g.text2 = "                               WITH EBITEN IT'S TOO EASY TO CREATE THIS KIND OF OLDSCHOOL DEMO, THIS IS ANOTHER NATIVE SCREEN BUILT WITH GOLANG AND EBITEN.                     WELL, IT'S NOT EXACTLY THE SAME EFFECTS, BUT IT'S CLOSE.             "
	g.text3 = "                                            YES... THERE IS A THIRD SCROLLTEXT IN THIS SCREEN..... MUCH LIKE ON THE ORIGINAL TCB FULLSCREEN DEMO, WE HAVE MULTIPLE DIFFERENT SCROLLERS... AND THESE ALL HAVE THIS COOL EFFECT ON THEM.... I HOPE YOU LIKE IT.....            "
	g.text4 = "                                                                               I GUESS WE SHOULD HAVE SOME GREETINGS, AS IT IS A DEMOSCREEN IN THE OLD SCHOOL STYLEE! SO HERE THEY ARE.... THE GREETZ GO OUT TO:  ALL MEMBERS OF DMA (PDM, COCO, JINX, CORWIN, DWORKIN) - MELLOW MAN - NONAMENO - THE UNION (MAD MAX FROM TEX FOR THE MUSIC) - COMMODOREBLOG - ELKMOOSE AND ANYONE ELSE I MAY HAVE MISSED!    LETZ WRAP..............       "

	return g
}

// loadImage charge une image depuis les assets embarqués
func (g *Game) loadImage(path string) (*ebiten.Image, error) {
	data, err := assets.ReadFile(path)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

// Init initialise toutes les ressources
func (g *Game) Init() error {
	var err error

	// Charger les images
	g.logoImg, err = g.loadImage("assets/logo.png")
	if err != nil {
		return fmt.Errorf("failed to load logo: %v", err)
	}

	g.titleImg, err = g.loadImage("assets/title.png")
	if err != nil {
		return fmt.Errorf("failed to load title: %v", err)
	}

	g.rasterImg, err = g.loadImage("assets/raster.png")
	if err != nil {
		return fmt.Errorf("failed to load raster: %v", err)
	}

	g.tileImg, err = g.loadImage("assets/tcb_tile.png")
	if err != nil {
		return fmt.Errorf("failed to load tile: %v", err)
	}

	g.fontImg, err = g.loadImage("assets/font.png")
	if err != nil {
		return fmt.Errorf("failed to load font: %v", err)
	}

	// Créer les canvas virtuels
	g.spriteCanvas = ebiten.NewImage(screenWidth/2, screenHeight/2)
	g.titleCanvas = ebiten.NewImage(528, 36)
	g.tileCanvas = ebiten.NewImage(canvasWidth, canvasHeight)

	// Initialiser le pattern de tiles
	for y := 0; y < canvasHeight; y += tileHeight {
		for x := 0; x < canvasWidth; x += tileWidth {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			g.tileCanvas.DrawImage(g.tileImg, op)
		}
	}

	// Initialiser l'audio
	g.audioContext = audio.NewContext(44100)

	// Charger la musique
	musicData, err := assets.ReadFile("assets/music.mp3")
	if err != nil {
		return fmt.Errorf("failed to load music: %v", err)
	}

	musicReader := bytes.NewReader(musicData)
	decodedMusic, err := mp3.DecodeWithSampleRate(44100, musicReader)
	if err != nil {
		return fmt.Errorf("failed to decode music: %v", err)
	}

	// Créer un loop infini
	loop := audio.NewInfiniteLoop(decodedMusic, decodedMusic.Length())
	g.audioPlayer, err = g.audioContext.NewPlayer(loop)
	if err != nil {
		return fmt.Errorf("failed to create audio player: %v", err)
	}

	// Démarrer la musique
	g.audioPlayer.Play()

	return nil
}

// mapCharToFont convertit un caractère en index de sprite de police
func mapCharToFont(charCode int) int {
	// Index 0 : Espace
	// Index 1-32 : ! " # $ % & ' ( ) * + , - . / 0-9 : ; < = > ? @
	// Index 33-58 : A-Z

	fontIndex := 0

	if charCode == 32 { // espace
		fontIndex = 0
	} else if charCode >= 33 && charCode <= 64 { // ! à @
		fontIndex = charCode - 32 // ! = index 1, @ = index 32
	} else if charCode >= 65 && charCode <= 90 { // A-Z
		fontIndex = (charCode - 65) + 33 // A = index 33, Z = index 58
	} else if charCode >= 97 && charCode <= 122 { // a-z -> A-Z
		fontIndex = (charCode - 97) + 33 // Convertir minuscules en majuscules
	} else {
		fontIndex = 0 // défaut = espace
	}

	// S'assurer que l'index est dans les limites (0-69 pour une grille 10x7)
	if fontIndex < 0 {
		fontIndex = 0
	}
	if fontIndex > 69 {
		fontIndex = 69
	}

	return fontIndex
}

// drawScroller dessine un texte défilant avec effet
func (g *Game) drawScroller(screen *ebiten.Image, textStr string, scrollX float64, scrollerID int, baseY float64) float64 {
	t := time.Since(g.startTime).Seconds() + 19
	sp := int(scrollX / 64)
	xs := math.Sin(t*0.25)*0.5 + 0.5
	xs = math.Sqrt(1 - xs*xs)

	for i := sp + 8; i >= sp; i-- {
		if i >= 0 && i < len(textStr) {
			z := math.Sin((t+float64(i)*0.15)*5)*0.5 + 1.5

			charCode := int(textStr[i])
			fontIndex := mapCharToFont(charCode)

			drawX := math.Floor((float64(i)*64 - 40 - math.Sin(t*7+float64(i)*18)*32*xs - scrollX) * 2)
			drawY := math.Floor(math.Sin((t+float64(i)*0.1)*7)*42*(math.Sin(t*0.5)*0.5+0.5) + baseY - z*32)

			var scale float64
			if scrollerID == 1 || scrollerID == 2 {
				scale = 3 - z
			} else {
				scale = z
			}

			// Test de visibilité
			if drawX > -100 && drawX < 900 && drawY > -100 && drawY < 700 && scale > 0.1 {
				// Calculer la position dans le sprite sheet de la font
				cols := 10 // 10 colonnes dans la font bitmap (10x7 = 70 caractères)
				srcX := (fontIndex % cols) * fontCharWidth
				srcY := (fontIndex / cols) * fontCharHeight

				// Dessiner le caractère
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(drawX, drawY)
				op.ColorM.Scale(1, 1, 1, 0.9)

				// Découper le caractère depuis la font bitmap
				charImg := g.fontImg.SubImage(image.Rect(srcX, srcY, srcX+fontCharWidth, srcY+fontCharHeight)).(*ebiten.Image)
				screen.DrawImage(charImg, op)
			}
		}
	}

	// Retourner la nouvelle position de scroll
	return math.Mod(scrollX+4, float64(len(textStr)*64))
}

// Update met à jour l'état du jeu
func (g *Game) Update() error {
	// Mise à jour des positions d'effet
	if g.fxFlag >= 1 {
		g.posXi += 0.008
	}
	if g.fxFlag >= 2 {
		g.posZi += 0.003
	}
	if g.fxFlag >= 3 {
		g.posRi += 0.005
	}

	// Gestion des phases d'effet
	if g.posXi >= 2 {
		g.fxFlag = 2
	}
	if g.posZi >= 1.5 {
		g.fxFlag = 3
	}

	// Animation d'intro
	if g.fxFlag == 0 {
		g.initX -= 4
		if g.initX <= -float64(screenWidth*2) {
			g.initR -= 1
		}
		if g.initR <= -45 {
			g.fxFlag = 1
		}
	}

	// Gestion du titre DMA
	if g.hold >= 1 {
		g.hold--
	}
	if g.hold <= 0 {
		g.logoX += 0.0125
	}

	// Animation des rasters
	g.rasterY1 -= 2
	g.rasterY2 -= 2
	if g.rasterY1 <= -72 {
		g.rasterY1 = 72
	}
	if g.rasterY2 <= -72 {
		g.rasterY2 = 72
	}

	// Incrémenter le compteur
	g.loopCounter++
	g.debugFrameCount++

	return nil
}

// Draw dessine le jeu
func (g *Game) Draw(screen *ebiten.Image) {
	// Fond noir
	screen.Fill(color.Black)

	// Dessiner l'effet de fond avec tiles
	if g.fxFlag >= 1 {
		zoom := 0.5 + math.Abs(math.Sin(g.posZi)*2.5)
		rotOriginal := 360 / 4 * math.Cos(g.posRi*4-math.Cos(g.posRi-0.01))
		rot := rotOriginal * 0.3 * math.Pi / 180 // Convertir en radians et réduire

		// Oscillations
		oscX := (float64(screenWidth) / 4) * math.Cos(g.posXi*4-math.Cos(g.posXi-0.1))
		oscY := (float64(screenHeight) / 2.7) * -math.Sin(g.posXi*2.3-math.Cos(g.posXi-0.1))

		centerX := float64(screenWidth)/2 + oscX
		centerY := float64(screenHeight)/2 + oscY

		// Dessiner le canvas de tiles avec transformation
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-float64(canvasWidth)/2, -float64(canvasHeight)/2)
		op.GeoM.Rotate(rot)
		op.GeoM.Scale(zoom, zoom)
		op.GeoM.Translate(centerX, centerY)
		screen.DrawImage(g.tileCanvas, op)
	} else {
		// Phase d'intro
		introX := g.initX + float64(screenWidth)/2
		introY := float64(270)
		introRot := g.initR * 0.3 * math.Pi / 180

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-float64(canvasWidth)/2, -float64(canvasHeight)/2)
		op.GeoM.Rotate(introRot)
		op.GeoM.Translate(introX, introY)
		screen.DrawImage(g.tileCanvas, op)
	}

	// Dessiner les scrollers
	g.scrollX1 = g.drawScroller(screen, g.text1, g.scrollX1, 1, 500)
	g.scrollX2 = g.drawScroller(screen, g.text2, g.scrollX2, 2, 250)
	g.scrollX3 = g.drawScroller(screen, g.text3, g.scrollX3, 3, 375)
	g.scrollX4 = g.drawScroller(screen, g.text4, g.scrollX4, 4, 125)

	// Masquer le haut de l'écran
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, 64, color.Black)

	// Dessiner le titre avec rasters
	g.titleCanvas.Fill(color.Black)

	// Dessiner les rasters
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(24, 1)
	op.GeoM.Translate(0, g.rasterY1)
	g.titleCanvas.DrawImage(g.rasterImg, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(24, 1)
	op.GeoM.Translate(0, g.rasterY2)
	g.titleCanvas.DrawImage(g.rasterImg, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(24, 1)
	op.GeoM.Translate(0, g.rasterY2+72)
	g.titleCanvas.DrawImage(g.rasterImg, op)

	// Dessiner le titre
	op = &ebiten.DrawImageOptions{}
	g.titleCanvas.DrawImage(g.titleImg, op)

	// Dessiner la surface du titre avec mouvement oscillant
	titleX := 64 + 768*math.Cos(g.logoX)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(titleX, 14)
	screen.DrawImage(g.titleCanvas, op)

	// Dessiner les logos animés (10 instances)
	g.spriteCanvas.Fill(color.Transparent)

	midX := float64(screenWidth/2-32) / 2 // Ajusté pour un sprite 32x32
	midY := 24 + float64(screenHeight/2-32)/2
	incY := float64(screenHeight/2-64) / 4 // Ajusté pour la nouvelle taille

	for s := 0; s < 10; s++ { // 10 sprites
		nit := float64(g.loopCounter + s*5)
		spX := midX + midX*math.Sin(nit/25)*math.Cos(nit/300)
		spY := midY + incY*math.Sin(nit/37) + incY*math.Cos(nit/17)

		// Dessiner le logo complet
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(spX, spY)
		g.spriteCanvas.DrawImage(g.logoImg, op)
	}

	// Dessiner la surface des sprites agrandie
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(2, 2)
	screen.DrawImage(g.spriteCanvas, op)
}

// Layout définit la taille de l'écran
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := NewGame()

	// Initialiser le jeu
	if err := game.Init(); err != nil {
		log.Fatal(err)
	}

	// Configurer la fenêtre
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("VIVA TCB! - Go/Ebiten Port")

	// Lancer le jeu
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// =============================================
// STRUCTURE DES ASSETS REQUIS :
// =============================================
// assets/
// ├── logo.png       (logo 32x32 pixels)
// ├── title.png      (titre "VIVA TCB")
// ├── raster.png     (effet de balayage)
// ├── tcb_tile.png   (texture de fond répétée)
// ├── font.png       (font bitmap avec caractères 42x40)
// └── music.mp3   (musique de fond)
//
// Note: Les assets doivent être placés dans un dossier "assets"
// à la racine du projet pour être embarqués avec go:embed
