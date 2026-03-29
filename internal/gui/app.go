package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	apptheme "github.com/havoczephyr/material-overlay-gui/internal/theme"
	"github.com/havoczephyr/material-overlay-gui/internal/service"
)

// App is the main application struct holding window and service references.
type App struct {
	fyneApp    fyne.App
	window     fyne.Window
	svc        *service.CardService
	searchBar  *widget.Entry
	contentBox *fyne.Container
	recentCards []string
}

// Run creates and launches the application.
func Run() error {
	svc, err := service.NewCardService()
	if err != nil {
		return err
	}

	a := &App{
		fyneApp: app.NewWithID("com.mhz.material-overlay"),
		svc:     svc,
	}

	a.fyneApp.Settings().SetTheme(&apptheme.OverlayTheme{})

	a.window = a.fyneApp.NewWindow("material(Overlay)")
	a.window.Resize(fyne.NewSize(960, 680))
	a.window.SetFixedSize(false)

	a.searchBar = widget.NewEntry()
	a.searchBar.SetPlaceHolder("Search cards...")
	a.searchBar.OnSubmitted = func(query string) {
		if query != "" {
			a.showSearch(query)
		}
	}

	a.contentBox = container.NewStack()

	toolbar := a.buildToolbar()
	layout := container.NewBorder(toolbar, nil, nil, nil, a.contentBox)
	a.window.SetContent(layout)

	// Keyboard shortcut: Cmd/Ctrl+F to focus search
	a.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierShortcutDefault},
		func(_ fyne.Shortcut) {
			a.window.Canvas().Focus(a.searchBar)
		},
	)

	// Load Genesys points in background
	go func() {
		_ = svc.LoadGenesysPoints()
	}()

	// Show home view
	a.showHome()

	a.window.ShowAndRun()
	return nil
}

func (a *App) buildToolbar() fyne.CanvasObject {
	brandLabel := widget.NewRichTextFromMarkdown("**material(Overlay)**")

	backBtn := widget.NewButton("< Back", func() {
		a.showHome()
	})

	searchBox := container.NewGridWrap(fyne.NewSize(300, 36), a.searchBar)

	return container.NewBorder(
		nil, nil,
		container.NewHBox(backBtn, brandLabel),
		searchBox,
	)
}

func (a *App) setContent(content fyne.CanvasObject) {
	fyne.Do(func() {
		a.contentBox.Objects = []fyne.CanvasObject{content}
		a.contentBox.Refresh()
	})
}

func (a *App) addRecentCard(name string) {
	// Remove if already present
	for i, n := range a.recentCards {
		if n == name {
			a.recentCards = append(a.recentCards[:i], a.recentCards[i+1:]...)
			break
		}
	}
	// Prepend
	a.recentCards = append([]string{name}, a.recentCards...)
	// Cap at 10
	if len(a.recentCards) > 10 {
		a.recentCards = a.recentCards[:10]
	}
}
