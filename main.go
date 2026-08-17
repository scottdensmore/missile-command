package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"missile-command/game"
)

func main() {
	ebiten.SetWindowSize(1024, 924)
	ebiten.SetWindowTitle("Missile Command (1980 Atari Arcade Replica)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g := game.NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
