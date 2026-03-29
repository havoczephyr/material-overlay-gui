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
	entries    []api.GalleryEntry
	thumbData  [][]byte // cached image data per entry
	thumbFrames []*fyne.Container
	idx        int
	mainImg    *fyne.Container
	infoLabel  *canvas.Text
	artLabel   *canvas.Text
	thumbStrip *fyne.Container
	prevBtn    *widget.Button
	nextBtn    *widget.Button
}

func (a *App) loadGallery(cardName string, box *fyne.Container) {
	entries, err := a.svc.FetchGalleryEntries(cardName)
	if err != nil || len(entries) == 0 {
		fyne.Do(func() {
			box.Objects = nil
			dimText := canvas.NewText("No gallery images available.", theme.ColorFGDim)
			dimText.TextSize = 13
			box.Add(dimText)
			box.Refresh()
		})
		return
	}

	gs := &galleryState{
		entries:   entries,
		thumbData: make([][]byte, len(entries)),
	}

	// Build all objects (safe to create off-thread; not yet visible)
	mainPlaceholder := placeholderImage()
	gs.mainImg = container.NewCenter(mainPlaceholder)

	gs.infoLabel = canvas.NewText(fmt.Sprintf("Artwork 1 of %d", len(entries)), theme.ColorFGDim)
	gs.infoLabel.TextSize = 13
	gs.infoLabel.Alignment = fyne.TextAlignCenter

	gs.artLabel = canvas.NewText(entries[0].Label, theme.ColorFG)
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
	gs.prevBtn.Disable()

	gs.nextBtn = widget.NewButton("Next >", func() {
		if gs.idx < len(gs.entries)-1 {
			gs.selectImage(gs.idx+1, a)
		}
	})
	if len(entries) <= 1 {
		gs.nextBtn.Disable()
	}

	navRow := container.NewBorder(nil, nil, gs.prevBtn, gs.nextBtn, gs.infoLabel)

	// Add to visible container on main thread
	fyne.Do(func() {
		bottomSection := container.NewVBox(gs.artLabel, thumbScroll, navRow)
		galleryLayout := container.NewBorder(nil, bottomSection, nil, nil, gs.mainImg)
		box.Objects = []fyne.CanvasObject{galleryLayout}
		box.Refresh()
	})

	// Load thumbnails sequentially in background (respects rate limiting)
	go gs.loadAllThumbnails(a)
}

// selectImage is called from tappable callbacks (main thread).
func (gs *galleryState) selectImage(idx int, a *App) {
	gs.idx = idx
	gs.updateLabels()
	gs.updateThumbHighlights()

	// Show the main image
	if gs.thumbData[idx] != nil {
		gs.setMainImage(gs.thumbData[idx], gs.entries[idx].Filename)
	} else {
		// Not yet loaded - fetch it then update on main thread
		go func() {
			data, err := a.svc.FetchGalleryImage(gs.entries[idx])
			if err != nil {
				return
			}
			gs.thumbData[idx] = data
			fyne.Do(func() {
				gs.setMainImage(data, gs.entries[idx].Filename)
			})
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
	img.SetMinSize(fyne.NewSize(200, 293))
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

		idx := i
		entryFilename := entry.Filename
		fyne.Do(func() {
			// Update thumbnail frame
			img := canvas.NewImageFromReader(bytes.NewReader(data), entryFilename)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(80, 117))

			frame := gs.thumbFrames[idx]
			frame.Objects = []fyne.CanvasObject{thumbnailFrame(img, idx == gs.idx)}
			frame.Refresh()

			// Set main image for the first entry
			if idx == 0 && gs.idx == 0 {
				gs.setMainImage(data, entryFilename)
			}
		})
	}
}
