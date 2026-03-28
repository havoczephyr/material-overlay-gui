package main

import (
	"log"

	"github.com/havoczephyr/material-overlay-gui/internal/gui"
)

func main() {
	if err := gui.Run(); err != nil {
		log.Fatal(err)
	}
}
