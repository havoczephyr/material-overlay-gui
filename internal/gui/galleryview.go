package gui

import (
	"bytes"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) loadGallery(cardName string, box *fyne.Container) {
	entries, err := a.svc.FetchGalleryEntries(cardName)
	if err != nil || len(entries) == 0 {
		box.Objects = nil
		dimText := canvas.NewText("No gallery images available.", theme.ColorFGDim)
		dimText.TextSize = 13
		box.Add(dimText)
		box.Refresh()
		return
	}

	idx := 0
	imgWidget := placeholderImage()
	imgContainer := container.NewCenter(imgWidget)

	infoLabel := canvas.NewText("", theme.ColorFGDim)
	infoLabel.TextSize = 13
	infoLabel.Alignment = fyne.TextAlignCenter

	artLabel := canvas.NewText("", theme.ColorFG)
	artLabel.TextSize = 14
	artLabel.Alignment = fyne.TextAlignCenter

	prevBtn := widget.NewButton("< Prev", nil)
	nextBtn := widget.NewButton("Next >", nil)

	updateGallery := func() {
		entry := entries[idx]
		infoLabel.Text = fmt.Sprintf("Artwork %d of %d", idx+1, len(entries))
		infoLabel.Refresh()
		artLabel.Text = entry.Label
		artLabel.Refresh()

		prevBtn.Disable()
		nextBtn.Disable()

		go func() {
			a.loadGalleryImage(entry, imgContainer)
			if idx > 0 {
				prevBtn.Enable()
			}
			if idx < len(entries)-1 {
				nextBtn.Enable()
			}
		}()
	}

	prevBtn.OnTapped = func() {
		if idx > 0 {
			idx--
			updateGallery()
		}
	}
	nextBtn.OnTapped = func() {
		if idx < len(entries)-1 {
			idx++
			updateGallery()
		}
	}

	navRow := container.NewBorder(nil, nil, prevBtn, nextBtn, infoLabel)

	box.Objects = nil
	box.Add(imgContainer)
	box.Add(artLabel)
	box.Add(navRow)
	box.Refresh()

	updateGallery()
}

func (a *App) loadGalleryImage(entry api.GalleryEntry, imgContainer *fyne.Container) {
	data, err := a.svc.FetchGalleryImage(entry)
	if err != nil {
		return
	}

	img := canvas.NewImageFromReader(bytes.NewReader(data), entry.Filename)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(300, 440))
	imgContainer.Objects = []fyne.CanvasObject{img}
	imgContainer.Refresh()
}
