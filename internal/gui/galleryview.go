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

// galleryState tracks the state of the thumbnail gallery.
type galleryState struct {
	entries      []api.GalleryEntry
	thumbData    [][]byte // cached image data per entry
	thumbFrames  []*fyne.Container
	idx          int
	mainImg      *fyne.Container
	infoLabel    *canvas.Text
	artLabel     *canvas.Text
	thumbStrip   *fyne.Container
	prevBtn      *widget.Button
	nextBtn      *widget.Button
}

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

	gs := &galleryState{
		entries:   entries,
		thumbData: make([][]byte, len(entries)),
	}

	// Main image display
	mainPlaceholder := placeholderImage()
	gs.mainImg = container.NewCenter(mainPlaceholder)

	// Info labels
	gs.infoLabel = canvas.NewText("", theme.ColorFGDim)
	gs.infoLabel.TextSize = 13
	gs.infoLabel.Alignment = fyne.TextAlignCenter

	gs.artLabel = canvas.NewText("", theme.ColorFG)
	gs.artLabel.TextSize = 14
	gs.artLabel.Alignment = fyne.TextAlignCenter

	// Build thumbnail strip
	gs.thumbFrames = make([]*fyne.Container, len(entries))
	thumbItems := make([]fyne.CanvasObject, len(entries))

	for i := range entries {
		idx := i
		placeholder := placeholderThumb()
		frame := container.NewStack(thumbnailFrame(placeholder, i == 0))
		gs.thumbFrames[i] = frame

		tappable := newTappableImage(frame, func() {
			gs.selectImage(idx, a)
		})
		thumbItems[i] = tappable
	}

	gs.thumbStrip = container.NewHBox(thumbItems...)
	thumbScroll := container.NewHScroll(gs.thumbStrip)
	thumbScroll.SetMinSize(fyne.NewSize(0, 135))

	// Navigation buttons
	gs.prevBtn = widget.NewButton("< Prev", func() {
		if gs.idx > 0 {
			gs.selectImage(gs.idx-1, a)
		}
	})
	gs.nextBtn = widget.NewButton("Next >", func() {
		if gs.idx < len(gs.entries)-1 {
			gs.selectImage(gs.idx+1, a)
		}
	})

	navRow := container.NewBorder(nil, nil, gs.prevBtn, gs.nextBtn, gs.infoLabel)

	// Assemble gallery layout
	box.Objects = nil
	box.Add(gs.mainImg)
	box.Add(gs.artLabel)
	box.Add(thumbScroll)
	box.Add(navRow)
	box.Refresh()

	// Update labels for first image
	gs.updateLabels()

	// Load thumbnails sequentially in background (respects rate limiting)
	go gs.loadAllThumbnails(a)
}

func (gs *galleryState) selectImage(idx int, a *App) {
	gs.idx = idx
	gs.updateLabels()
	gs.updateThumbHighlights()

	// Show the main image
	if gs.thumbData[idx] != nil {
		gs.setMainImage(gs.thumbData[idx], gs.entries[idx].Filename)
	} else {
		// Not yet loaded - load it now
		go func() {
			data, err := a.svc.FetchGalleryImage(gs.entries[idx])
			if err != nil {
				return
			}
			gs.thumbData[idx] = data
			gs.setMainImage(data, gs.entries[idx].Filename)
		}()
	}
}

func (gs *galleryState) updateLabels() {
	gs.infoLabel.Text = fmt.Sprintf("Artwork %d of %d", gs.idx+1, len(gs.entries))
	gs.infoLabel.Refresh()
	gs.artLabel.Text = gs.entries[gs.idx].Label
	gs.artLabel.Refresh()

	gs.prevBtn.Enable()
	gs.nextBtn.Enable()
	if gs.idx == 0 {
		gs.prevBtn.Disable()
	}
	if gs.idx >= len(gs.entries)-1 {
		gs.nextBtn.Disable()
	}
}

func (gs *galleryState) updateThumbHighlights() {
	for i, frame := range gs.thumbFrames {
		var content fyne.CanvasObject
		if gs.thumbData[i] != nil {
			img := canvas.NewImageFromReader(bytes.NewReader(gs.thumbData[i]), gs.entries[i].Filename)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(80, 117))
			content = img
		} else {
			content = placeholderThumb()
		}
		frame.Objects = []fyne.CanvasObject{thumbnailFrame(content, i == gs.idx)}
		frame.Refresh()
	}
}

func (gs *galleryState) setMainImage(data []byte, name string) {
	img := canvas.NewImageFromReader(bytes.NewReader(data), name)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(300, 440))
	gs.mainImg.Objects = []fyne.CanvasObject{img}
	gs.mainImg.Refresh()
}

// loadAllThumbnails fetches each gallery image sequentially.
// Each FetchGalleryImage call goes through the Yugipedia rate limiter (1req/sec).
func (gs *galleryState) loadAllThumbnails(a *App) {
	for i, entry := range gs.entries {
		data, err := a.svc.FetchGalleryImage(entry)
		if err != nil {
			continue
		}
		gs.thumbData[i] = data

		// Update thumbnail frame
		img := canvas.NewImageFromReader(bytes.NewReader(data), entry.Filename)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(80, 117))

		frame := gs.thumbFrames[i]
		frame.Objects = []fyne.CanvasObject{thumbnailFrame(img, i == gs.idx)}
		frame.Refresh()

		// Set main image for the first entry
		if i == 0 && gs.idx == 0 {
			gs.setMainImage(data, entry.Filename)
		}
	}
}
