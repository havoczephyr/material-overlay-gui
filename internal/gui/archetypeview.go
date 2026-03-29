package gui

import (
	"bytes"
	"fmt"
	"image/color"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showArchetype(name string) {
	loading := container.NewCenter(widget.NewLabel("Loading " + name + "..."))
	a.setContent(loading)

	go func() {
		var (
			article string
			imgData []byte
			cards   []api.YGOProCard
			artErr  error
			cardErr error
			wg      sync.WaitGroup
		)

		wg.Add(3)

		go func() {
			defer wg.Done()
			article, artErr = a.svc.FetchArchetypeArticle(name)
		}()

		go func() {
			defer wg.Done()
			imgData, _ = a.svc.FetchArchetypeSplashImage(name)
		}()

		go func() {
			defer wg.Done()
			cards, cardErr = a.svc.FetchArchetypeCards(name)
		}()

		wg.Wait()
		a.renderArchetypeView(name, article, artErr, imgData, cards, cardErr)
	}()
}

func (a *App) renderArchetypeView(name, article string, artErr error, imgData []byte, cards []api.YGOProCard, cardErr error) {
	// ── Article section (compact upper-left header + scrollable article) ──
	var imgWidget fyne.CanvasObject
	if imgData != nil {
		img := canvas.NewImageFromReader(bytes.NewReader(imgData), name)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(85, 125))
		imgWidget = img
	} else {
		rect := canvas.NewRectangle(theme.ColorBGLight)
		rect.SetMinSize(fyne.NewSize(85, 125))
		rect.CornerRadius = 8
		imgWidget = rect
	}

	// Fade overlay sits on top of the image — starts transparent, fades to BG color on scroll
	bgR, bgG, bgB, _ := theme.ColorBG.RGBA()
	fadeOverlay := canvas.NewRectangle(color.NRGBA{R: uint8(bgR >> 8), G: uint8(bgG >> 8), B: uint8(bgB >> 8), A: 0})
	fadeOverlay.SetMinSize(fyne.NewSize(85, 125))
	imgWithFade := container.NewStack(imgWidget, fadeOverlay)

	nameText := canvas.NewText(name, theme.ColorPrimary)
	nameText.TextSize = 14
	nameText.TextStyle.Bold = true

	subtitleText := canvas.NewText("Archetype", theme.ColorFGDim)
	subtitleText.TextSize = 12

	// Extract color components for scroll fade callback
	pR, pG, pB, _ := theme.ColorPrimary.RGBA()
	dimR, dimG, dimB, _ := theme.ColorFGDim.RGBA()

	var articleWidget fyne.CanvasObject
	if artErr != nil {
		articleWidget = widget.NewLabel("Failed to load article: " + artErr.Error())
	} else if article == "" {
		dimText := canvas.NewText("No article available.", theme.ColorFGDim)
		dimText.TextSize = 13
		articleWidget = dimText
	} else {
		label := widget.NewLabel(article)
		label.Wrapping = fyne.TextWrapWord
		label.Selectable = true
		articleWidget = label
	}

	// Compact header: thumbnail left, title + subtitle stacked right
	header := container.NewHBox(
		imgWithFade,
		container.NewVBox(nameText, subtitleText),
	)

	articleScroll := container.NewVScroll(container.NewPadded(articleWidget))

	// Scroll-based fade: title and thumbnail decrease in opacity as user scrolls down
	const fadeDistance = float32(150)
	articleScroll.OnScrolled = func(pos fyne.Position) {
		ratio := pos.Y / fadeDistance
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}

		// Image fade overlay: transparent → opaque background color
		overlayAlpha := uint8(ratio * 255)
		fadeOverlay.FillColor = color.NRGBA{R: uint8(bgR >> 8), G: uint8(bgG >> 8), B: uint8(bgB >> 8), A: overlayAlpha}
		fadeOverlay.Refresh()

		// Title text: fully visible → invisible
		textAlpha := uint8((1 - ratio) * 255)
		nameText.Color = color.NRGBA{R: uint8(pR >> 8), G: uint8(pG >> 8), B: uint8(pB >> 8), A: textAlpha}
		nameText.Refresh()

		subtitleText.Color = color.NRGBA{R: uint8(dimR >> 8), G: uint8(dimG >> 8), B: uint8(dimB >> 8), A: textAlpha}
		subtitleText.Refresh()
	}

	articleSection := container.NewBorder(
		header, nil, nil, nil,
		articleScroll,
	)

	// ── Cards section ──
	cardHeader := sectionHeader(fmt.Sprintf("Cards in %s", name))
	cardListBox := container.NewVBox()

	if cardErr != nil {
		cardListBox.Add(widget.NewLabel("Failed to load cards: " + cardErr.Error()))
	} else if len(cards) == 0 {
		dimText := canvas.NewText("No cards found.", theme.ColorFGDim)
		dimText.TextSize = 13
		cardListBox.Add(dimText)
	} else {
		for _, c := range cards {
			cardListBox.Add(a.archetypeCardRow(c))
		}
	}

	cardSection := container.NewVBox(cardHeader, cardListBox)
	cardScroll := container.NewVScroll(container.NewPadded(cardSection))

	// ── Two-layer Stack (same pattern as card view gallery animation) ──
	// Fyne doesn't clip children, so when GridWrap shrinks to 0 the article
	// content overflows. Fix: article sits on a bottom layer at fixed height,
	// and the cards layer has an animated spacer + opaque bg that slides over it.

	splitHeight := float32(300)
	expandedHeight := float32(650)
	collapsedHeight := float32(0)
	currentHeight := splitHeight
	state := 0 // 0=split, 1=article expanded, 2=cards expanded

	// Bottom layer: article section always at full height, never animated
	articleFixedWrapper := container.NewGridWrap(fyne.NewSize(960, expandedHeight), articleSection)
	articleLayer := container.NewBorder(articleFixedWrapper, nil, nil, nil)

	// Top layer spacer: animated, pushes cards section down
	topSpacer := container.NewGridWrap(fyne.NewSize(960, splitHeight), layout.NewSpacer())

	var currentAnim *fyne.Animation
	animateTo := func(target float32) {
		if currentAnim != nil {
			currentAnim.Stop()
		}
		from := currentHeight
		currentAnim = canvas.NewSizeAnimation(
			fyne.NewSize(960, from),
			fyne.NewSize(960, target),
			time.Millisecond*300,
			func(s fyne.Size) {
				topSpacer.Layout = layout.NewGridWrapLayout(s)
				topSpacer.Refresh()
				a.contentBox.Refresh()
			},
		)
		currentAnim.Curve = fyne.AnimationEaseInOut
		currentAnim.Start()
		currentHeight = target
	}

	downBtn := widget.NewButton("▼", nil)
	upBtn := widget.NewButton("▲", nil)

	downBtn.OnTapped = func() {
		switch state {
		case 0: // split → article expanded
			animateTo(expandedHeight)
			state = 1
		case 2: // cards expanded → split
			animateTo(splitHeight)
			state = 0
		}
	}
	upBtn.OnTapped = func() {
		switch state {
		case 0: // split → cards expanded
			animateTo(collapsedHeight)
			state = 2
		case 1: // article expanded → split
			animateTo(splitHeight)
			state = 0
		}
	}

	toggleBg := canvas.NewRectangle(theme.ColorBGLight)
	toggleBg.SetMinSize(fyne.NewSize(0, 32))
	toggleBar := container.NewStack(
		toggleBg,
		container.NewCenter(container.NewHBox(downBtn, upBtn)),
	)

	// Top layer: opaque bg covers article beneath as cards slide up
	cardsBg := canvas.NewRectangle(theme.ColorBG)
	cardsWithBg := container.NewStack(cardsBg, container.NewBorder(toggleBar, nil, nil, nil, cardScroll))
	cardsLayer := container.NewBorder(topSpacer, nil, nil, nil, cardsWithBg)

	// Stack: article on bottom, cards slide over it on top
	fullView := container.NewStack(articleLayer, cardsLayer)

	a.setContent(fullView)
}

func (a *App) archetypeCardRow(c api.YGOProCard) fyne.CanvasObject {
	cardName := c.Name
	nameBtn := widget.NewButton(cardName, func() {
		a.showCardByName(cardName)
	})
	nameBtn.Importance = widget.LowImportance
	nameBtn.Alignment = widget.ButtonAlignLeading

	typeInfo := c.Type
	if c.Race != "" {
		typeInfo = c.Race + " / " + c.Type
	}

	var details []string
	if c.Attribute != "" {
		details = append(details, c.Attribute)
	}
	details = append(details, typeInfo)

	detailText := canvas.NewText(strings.Join(details, "  •  "), theme.ColorFGDim)
	detailText.TextSize = 12

	return container.NewVBox(newTappableButton(nameBtn), detailText)
}
