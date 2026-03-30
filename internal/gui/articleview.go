package gui

import (
	"bytes"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/card"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showArticle(title string) {
	loading := container.NewCenter(widget.NewLabel("Loading " + title + "..."))
	a.setContent(loading)

	go func() {
		var (
			segments []card.ContentSegment
			imgData  []byte
			artErr   error
			wg       sync.WaitGroup
		)

		wg.Add(2)

		go func() {
			defer wg.Done()
			segments, artErr = a.svc.FetchGenericArticle(title)
		}()

		go func() {
			defer wg.Done()
			imgData, _ = a.svc.FetchArticleSplashImage(title)
		}()

		wg.Wait()
		a.renderArticleView(title, segments, artErr, imgData)
	}()
}

func (a *App) renderArticleView(title string, segments []card.ContentSegment, artErr error, splashImg []byte) {
	// ── Parallax fade header (same pattern as archetype view) ──
	var imgWidget fyne.CanvasObject
	if splashImg != nil {
		img := canvas.NewImageFromReader(bytes.NewReader(splashImg), title)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(85, 125))
		imgWidget = img
	} else {
		rect := canvas.NewRectangle(theme.ColorBGLight)
		rect.SetMinSize(fyne.NewSize(85, 125))
		rect.CornerRadius = 8
		imgWidget = rect
	}

	bgR, bgG, bgB, _ := theme.ColorBG.RGBA()
	fadeOverlay := canvas.NewRectangle(color.NRGBA{R: uint8(bgR >> 8), G: uint8(bgG >> 8), B: uint8(bgB >> 8), A: 0})
	fadeOverlay.SetMinSize(fyne.NewSize(85, 125))
	imgWithFade := container.NewStack(imgWidget, fadeOverlay)

	nameText := canvas.NewText(title, theme.ColorPrimary)
	nameText.TextSize = 14
	nameText.TextStyle.Bold = true

	subtitleText := canvas.NewText("Wiki Article", theme.ColorFGDim)
	subtitleText.TextSize = 12

	pR, pG, pB, _ := theme.ColorPrimary.RGBA()
	dimR, dimG, dimB, _ := theme.ColorFGDim.RGBA()

	header := container.NewHBox(
		imgWithFade,
		container.NewVBox(nameText, subtitleText),
	)
	headerLayer := container.NewBorder(header, nil, nil, nil)

	// ── Article body ──
	var articleWidget fyne.CanvasObject
	if artErr != nil {
		articleWidget = widget.NewLabel("Failed to load article: " + artErr.Error())
	} else if len(segments) == 0 {
		dimText := canvas.NewText("No article available.", theme.ColorFGDim)
		dimText.TextSize = 13
		articleWidget = dimText
	} else {
		articleWidget = a.buildArticleContent(segments)
	}

	headerSpacer := container.NewGridWrap(fyne.NewSize(0, 135), layout.NewSpacer())
	articleContent := container.NewVBox(headerSpacer, container.NewPadded(articleWidget))
	articleScroll := container.NewVScroll(articleContent)

	// Scroll-based fade
	const fadeDistance = float32(135)
	articleScroll.OnScrolled = func(pos fyne.Position) {
		ratio := pos.Y / fadeDistance
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}

		overlayAlpha := uint8(ratio * 255)
		fadeOverlay.FillColor = color.NRGBA{R: uint8(bgR >> 8), G: uint8(bgG >> 8), B: uint8(bgB >> 8), A: overlayAlpha}
		fadeOverlay.Refresh()

		textAlpha := uint8((1 - ratio) * 255)
		nameText.Color = color.NRGBA{R: uint8(pR >> 8), G: uint8(pG >> 8), B: uint8(pB >> 8), A: textAlpha}
		nameText.Refresh()

		subtitleText.Color = color.NRGBA{R: uint8(dimR >> 8), G: uint8(dimG >> 8), B: uint8(dimB >> 8), A: textAlpha}
		subtitleText.Refresh()
	}

	fullView := container.NewStack(headerLayer, articleScroll)
	a.setContent(fullView)
}

// buildArticleContent creates a VBox with interleaved text labels and image placeholders.
func (a *App) buildArticleContent(segments []card.ContentSegment) fyne.CanvasObject {
	content := container.NewVBox()

	for _, seg := range segments {
		if seg.Image != nil {
			// Image placeholder — loaded asynchronously
			w := float32(seg.Image.Width)
			if w <= 0 {
				w = 300
			}
			h := w * 0.66 // reasonable aspect ratio default

			placeholder := canvas.NewRectangle(theme.ColorBGLight)
			placeholder.SetMinSize(fyne.NewSize(w, h))
			placeholder.CornerRadius = 4

			imgContainer := container.NewStack(placeholder)
			content.Add(container.NewPadded(imgContainer))

			// Async image load
			filename := seg.Image.Filename
			caption := seg.Image.Caption
			imgWidth := w
			go func() {
				data, err := a.svc.FetchArticleImage(filename)
				if err != nil {
					return
				}
				fyne.Do(func() {
					img := canvas.NewImageFromReader(bytes.NewReader(data), filename)
					img.FillMode = canvas.ImageFillContain
					img.SetMinSize(fyne.NewSize(imgWidth, imgWidth*0.66))

					var obj fyne.CanvasObject = img
					if caption != "" {
						capText := canvas.NewText(caption, theme.ColorFGDim)
						capText.TextSize = 12
						capText.Alignment = fyne.TextAlignCenter
						obj = container.NewVBox(img, capText)
					}
					imgContainer.Objects = []fyne.CanvasObject{obj}
					imgContainer.Refresh()
				})
			}()
		} else if seg.Text != "" {
			label := widget.NewLabel(seg.Text)
			label.Wrapping = fyne.TextWrapWord
			label.Selectable = true
			content.Add(label)
		}
	}

	if len(content.Objects) == 0 {
		dimText := canvas.NewText("No article available.", theme.ColorFGDim)
		dimText.TextSize = 13
		return dimText
	}

	return content
}


