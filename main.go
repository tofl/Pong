package main

import (
	_ "embed"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"
)

const (
	screenWidth  = 640
	screenHeight = 480
	gameTitle    = "PONG"
)

var (
	mainTitleFontFace font.Face
	subtitleFontFace  font.Face
	textBodyFontFace  font.Face
	gameStage         int8
	player1           Player
	player2           Player
	ball              Ball
	nextSideToPlay    string
	lastRoundTime     time.Time
)

//go:embed assets/retro-gaming.ttf
var retroGamingFontBytes []byte

//go:embed assets/pong-score.ttf
var pongScoreFontBytes []byte

// Text as displayed in the game
type Text struct {
	Text      string
	FontFace  font.Face
	PositionX int
	PositionY int
	Color     color.RGBA
}

func (t *Text) Draw(screen *ebiten.Image) {
	text.Draw(screen, t.Text, t.FontFace, t.PositionX, t.PositionY, t.Color)
}

// Players and their bouncers
type Player struct {
	Score         int8
	TextScore     Text
	BouncerHeight float64
	BouncerWidth  float64
	PositionX     float64
	PositionY     float64
}

func (p *Player) Draw(screen *ebiten.Image) {
	for x := 0.0; x < p.BouncerWidth; x++ {
		for y := 0.0; y < p.BouncerHeight; y++ {
			screen.Set(int(p.PositionX+x), int(p.PositionY+y), color.RGBA{200, 200, 200, 255})
		}
	}
}

func (p *Player) Move(up, down ebiten.Key) {
	if ebiten.IsKeyPressed(up) {
		if p.PositionY <= 10 {
			return
		}
		p.PositionY -= 6
	} else if ebiten.IsKeyPressed(down) {
		if p.PositionY >= screenHeight-p.BouncerHeight-10 {
			return
		}
		p.PositionY += 6
	}
}

type Ball struct {
	Width     float64
	PositionX float64
	PositionY float64
	Angle     float64
	Speed     float64
	Bounces   int
}

func (b *Ball) Draw(screen *ebiten.Image) {
	for x := 0.0; x < b.Width; x++ {
		for y := 0.0; y < b.Width; y++ {
			screen.Set(int(x+b.PositionX), int(y+b.PositionY), color.RGBA{200, 200, 200, 255})
		}
	}
}

func (b *Ball) Touch() {
	// Reaction to the ball hitting walls
	if b.PositionY <= 10 || b.PositionY > screenHeight-b.Width-10 {
		b.Angle = -b.Angle
	}

	// Reaction to the ball hitting the bouncer
	// Left bouncer
	if b.PositionX <= player1.PositionX+player1.BouncerWidth && b.PositionX+b.Width >= player1.PositionX {
		if b.PositionY < player1.PositionY && b.PositionY+ball.Width >= player1.PositionY {
			xDepth := player1.PositionX + player1.BouncerWidth - b.PositionX
			yDepth := b.PositionY + b.Width - player1.PositionY

			if xDepth > yDepth {
				b.Angle = -b.Angle
			} else {
				b.VerticalBounce(player1.PositionX + player1.BouncerWidth)
			}
		} else if b.PositionY+b.Width > player1.PositionY+player1.BouncerHeight && b.PositionY < player1.PositionY+player1.BouncerHeight {
			xDepth := player1.PositionX + player1.BouncerWidth - b.PositionX
			yDepth := player1.PositionY + player1.BouncerHeight - b.PositionY

			if xDepth > yDepth {
				b.Angle = -b.Angle
			} else {
				b.VerticalBounce(player1.PositionX + player1.BouncerWidth)
			}
		} else if b.PositionY+b.Width > player1.PositionY && b.PositionY < player1.PositionY+player1.BouncerHeight {
			b.VerticalBounce(player1.PositionX + player1.BouncerWidth)
		}
	}

	// Right bouncer
	if b.PositionX+b.Width >= player2.PositionX && b.PositionX < player2.PositionX+player2.BouncerWidth {
		if b.PositionY < player2.PositionY && b.PositionY+b.Width >= player2.PositionY {
			xDepth := b.PositionX + b.Width - player2.PositionX
			yDepth := b.PositionY + b.Width - player2.PositionY

			if xDepth > yDepth {
				b.Angle = -b.Angle
			} else {
				b.VerticalBounce(player2.PositionX - b.Width)
			}
		} else if b.PositionY <= player2.PositionY+player2.BouncerHeight && b.PositionY+b.Width > player2.PositionY+player2.BouncerHeight {
			xDepth := b.PositionX + b.Width - player2.PositionX
			yDepth := player2.PositionY + player2.BouncerHeight - b.PositionY

			if xDepth > yDepth {
				b.Angle = -b.Angle
			} else {
				b.VerticalBounce(player2.PositionX - b.Width)
			}
		} else if b.PositionY+b.Width > player2.PositionY && b.PositionY < player2.PositionY+player2.BouncerHeight {
			b.VerticalBounce(player2.PositionX - b.Width)
		}
	}
}

func (b *Ball) VerticalBounce(newPosition float64) {
	b.Angle = 180 - b.Angle - 10 + 20*rand.Float64()
	b.PositionX = newPosition
	b.Bounces++
	if b.Bounces == 1 {
		b.Speed = 7
	}
}

func (b *Ball) CheckGoals() {
	var winner *Player
	if b.PositionX+b.Width <= 0 {
		winner = &player2
	} else if b.PositionX >= screenWidth {
		winner = &player1
	}

	if winner != nil {
		winner.Score++
		winner.TextScore.Text = strconv.Itoa(int(winner.Score))
		b.Initialize()

		if winner.Score == 10 {
			gameStage = 3
		}
	}
}

func (b *Ball) Initialize() {
	var angle float64
	if nextSideToPlay == "right" {
		angle = -45 + float64(rand.Intn(91))
		nextSideToPlay = "left"
	} else {
		angle = 135 + float64(rand.Intn(91))
		nextSideToPlay = "right"
	}

	b.Width = 10
	b.Angle = angle
	b.Bounces = 0
	b.Speed = 3
	b.PositionX = (screenWidth - ball.Width) / 2
	b.PositionY = (screenHeight - ball.Width) / 2

	lastRoundTime = time.Now()
}

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) printMenu(screen *ebiten.Image) {
	// Main title
	mainTitle := Text{
		Text:      "PONG",
		FontFace:  mainTitleFontFace,
		PositionY: 80,
		Color:     color.RGBA{200, 200, 200, 255},
	}
	mainTitle.PositionX = (screenWidth - font.MeasureString(mainTitle.FontFace, mainTitle.Text).Round()) / 2
	mainTitle.Draw(screen)

	// Start instructions
	startInstructions := Text{
		Text:      "HIT P TO PLAY OR PAUSE",
		FontFace:  textBodyFontFace,
		PositionY: 120,
		Color:     color.RGBA{200, 200, 200, 255},
	}
	startInstructions.PositionX = (screenWidth - font.MeasureString(startInstructions.FontFace, startInstructions.Text).Round()) / 2
	startInstructions.Draw(screen)

	// Player 1 info
	playerOne := Text{
		Text:      "PLAYER 1\nUse D & E keys",
		FontFace:  textBodyFontFace,
		PositionY: 180,
		Color:     color.RGBA{200, 200, 200, 255},
	}
	playerOne.PositionX = (screenWidth - 2*font.MeasureString(playerOne.FontFace, playerOne.Text).Round()) / 4
	playerOne.Draw(screen)

	// Player 2 info
	playerTwo := Text{
		Text:      "PLAYER 2\nUse Arrow keys",
		FontFace:  textBodyFontFace,
		PositionY: 180,
		Color:     color.RGBA{200, 200, 200, 255},
	}
	playerTwo.PositionX = screenWidth/2 + (screenWidth-2*font.MeasureString(playerTwo.FontFace, playerTwo.Text).Round())/4
	playerTwo.Draw(screen)
}

func (g *Game) drawField(screen *ebiten.Image) {
	for x := 0; x <= screenWidth; x++ {
		for y := 0; y <= 10; y++ {
			screen.Set(x, y, color.RGBA{200, 200, 200, 255})
			screen.Set(x, y+screenHeight-10, color.RGBA{200, 200, 200, 255})
		}
	}

	// Draw delimiter line
	for squareCount, y := 0, 15; squareCount < 30; squareCount++ {
		for w := 0; w < 7; w++ {
			for h := 0; h < 7; h++ {
				screen.Set((screenWidth/2)-5+w, h+y, color.RGBA{200, 200, 200, 255})
			}
		}
		y += 17
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch gameStage {
	case 1: // Main menu
		g.printMenu(screen)
		if inpututil.IsKeyJustReleased(ebiten.KeyP) {
			gameStage = 2
		}
		break
	case 2: // Play
		if inpututil.IsKeyJustReleased(ebiten.KeyP) {
			// Pause the game
			gameStage = 4
		}

		// Draw field
		g.drawField(screen)

		// Show the score
		player1.TextScore.Draw(screen)
		player2.TextScore.Draw(screen)

		// Display and move bouncers
		player1.Move(ebiten.KeyE, ebiten.KeyD)
		player2.Move(ebiten.KeyUp, ebiten.KeyDown)

		player1.Draw(screen)
		player2.Draw(screen)

		if time.Since(lastRoundTime).Milliseconds() <= 1000 {
			return
		}

		// Display and move ball
		ball.Draw(screen)
		ball.PositionX = ball.PositionX + ball.Speed*math.Cos(float64(ball.Angle)*math.Pi/180)
		ball.PositionY = ball.PositionY + ball.Speed*math.Sin(float64(ball.Angle)*math.Pi/180)

		// Makes the ball bounce
		ball.Touch()

		// Look for goals
		ball.CheckGoals()
		break
	case 3: // Game finished
		g.drawField(screen)

		player1.Draw(screen)
		player1.TextScore.Draw(screen)

		player2.Draw(screen)
		player2.TextScore.Draw(screen)

		winnerAnnouncement := Text{
			FontFace:  subtitleFontFace,
			PositionY: 100,
			Color:     color.RGBA{200, 200, 200, 255},
		}

		newGameText := "HIT P TO PLAY AGAIN"
		newGame := Text{
			Text:      newGameText,
			FontFace:  textBodyFontFace,
			PositionY: 140,
			Color:     color.RGBA{200, 200, 200, 255},
		}

		if player1.Score > player2.Score {
			winnerAnnouncementText := "Player 1 wins!"
			winnerAnnouncement.Text = winnerAnnouncementText
			winnerAnnouncement.PositionX = (screenWidth / 4) - font.MeasureString(subtitleFontFace, winnerAnnouncementText).Round()/2
			newGame.PositionX = (screenWidth / 4) - font.MeasureString(textBodyFontFace, newGameText).Round()/2
		} else {
			winnerAnnouncementText := "Player 2 wins!"
			winnerAnnouncement.Text = winnerAnnouncementText
			winnerAnnouncement.PositionX = (3 * screenWidth / 4) - font.MeasureString(subtitleFontFace, winnerAnnouncementText).Round()/2
			newGame.PositionX = (3 * screenWidth / 4) - font.MeasureString(textBodyFontFace, newGameText).Round()/2
		}

		winnerAnnouncement.Draw(screen)
		newGame.Draw(screen)

		if inpututil.IsKeyJustReleased(ebiten.KeyP) {
			gameStage = 2
			player1.Score = 0
			player1.TextScore.Text = "0"

			player2.Score = 0
			player2.TextScore.Text = "0"

			ball.Initialize()
		}
		break
	case 4: // The game is paused
		g.drawField(screen)

		player1.Draw(screen)
		player1.TextScore.Draw(screen)

		player2.Draw(screen)
		player2.TextScore.Draw(screen)

		ball.Draw(screen)

		// Show instructions on the screen
		resumeGameText := "HIT P TO RESUME GAME"
		resumeGame := Text{
			Text:      resumeGameText,
			FontFace:  subtitleFontFace,
			PositionX: (screenWidth - font.MeasureString(subtitleFontFace, resumeGameText).Round()) / 2,
			PositionY: 100,
			Color:     color.RGBA{20, 20, 20, 255},
		}

		textBounds, _ := font.BoundString(subtitleFontFace, resumeGameText)

		textWidth := textBounds.Max.X.Round() - textBounds.Min.X.Round()
		textHeight := textBounds.Max.Y.Round() - textBounds.Min.Y.Round()

		vector.DrawFilledRect(
			screen,
			float32((screenWidth-font.MeasureString(subtitleFontFace, resumeGameText).Round())/2)-4,
			100-float32(textHeight)-2,
			float32(textWidth)+8,
			float32(textHeight+4),
			color.RGBA{200, 200, 200, 255},
			true,
		)
		resumeGame.Draw(screen)

		if inpututil.IsKeyJustReleased(ebiten.KeyP) {
			gameStage = 2
		}
		break
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle(gameTitle)

	// Load fonts
	textFont, err := truetype.Parse(retroGamingFontBytes)
	if err != nil {
		log.Fatal(err)
	}

	pongScoreFont, err := truetype.Parse(pongScoreFontBytes)
	if err != nil {
		log.Fatal("oh no", err)
	}

	// Create font faces
	mainTitleFontFace = truetype.NewFace(textFont, &truetype.Options{
		Size: 80,
		DPI:  72,
	})

	subtitleFontFace = truetype.NewFace(textFont, &truetype.Options{
		Size: 22,
		DPI:  72,
	})

	textBodyFontFace = truetype.NewFace(textFont, &truetype.Options{
		Size: 12,
		DPI:  72,
	})

	pongScoreFontFace := truetype.NewFace(pongScoreFont, &truetype.Options{
		Size: 44,
		DPI:  72,
	})

	// Initialize the game
	gameStage = 1

	// Initialize the players
	player1 = Player{
		Score:         0,
		BouncerHeight: 50,
		BouncerWidth:  10,
		PositionX:     15,
		PositionY:     0,
	}
	player1.PositionY = (screenHeight - player1.BouncerHeight) / 2
	player1.TextScore = Text{
		Text:      strconv.Itoa(int(player1.Score)),
		FontFace:  pongScoreFontFace,
		PositionX: screenWidth/2 - 50,
		PositionY: 60,
		Color:     color.RGBA{200, 200, 200, 255},
	}
	player1.TextScore.PositionX = (screenWidth / 2) - 70 - font.MeasureString(player1.TextScore.FontFace, player1.TextScore.Text).Round()

	player2 = Player{
		Score:         0,
		BouncerHeight: 50,
		BouncerWidth:  10,
		PositionX:     screenWidth - 25,
		PositionY:     0,
	}
	player2.PositionY = (screenHeight - player2.BouncerHeight) / 2
	player2.TextScore = Text{
		Text:      strconv.Itoa(int(player2.Score)),
		FontFace:  pongScoreFontFace,
		PositionY: 60,
		Color:     color.RGBA{200, 200, 200, 255},
	}
	player2.TextScore.PositionX = (screenWidth / 2) + 70

	// Initialize the ball
	if rand.Intn(2) == 0 {
		nextSideToPlay = "right"
	} else {
		nextSideToPlay = "left"
	}

	ball.Initialize()

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
