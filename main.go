package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"missile-command/assets"
	"missile-command/game"
)

func main() {
	ebiten.SetWindowSize(1024, 924)
	ebiten.SetWindowTitle("Missile Command (1980 Atari Arcade Replica)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowIcon(assets.GetWindowIcons())

	g := game.NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
