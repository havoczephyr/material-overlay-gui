package gui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showSearch(query string) {
	header := sectionHeader("Search Results")
	statusLabel := canvas.NewText("Searching for \""+query+"\"...", theme.ColorFGDim)
	statusLabel.TextSize = 13

	// Three result columns
	cardResultsList := container.NewVBox()
	archetypeResultsList := container.NewVBox()
	articleResultsList := container.NewVBox()

	cardColumn := container.NewVBox(sectionHeader("Cards"), cardResultsList)
	archColumn := container.NewVBox(sectionHeader("Archetypes"), archetypeResultsList)
	articleColumn := container.NewVBox(sectionHeader("Wiki Articles"), articleResultsList)

	// Three-column layout: Cards | Archetypes | Wiki Articles
	rightSplit := container.NewHSplit(
		container.NewVScroll(container.NewPadded(archColumn)),
		container.NewVScroll(container.NewPadded(articleColumn)),
	)
	rightSplit.SetOffset(0.5)

	resultsSplit := container.NewHSplit(
		container.NewVScroll(container.NewPadded(cardColumn)),
		rightSplit,
	)
	resultsSplit.SetOffset(0.4)

	content := container.NewBorder(
		container.NewPadded(container.NewVBox(header, statusLabel)),
		nil, nil, nil,
		resultsSplit,
	)
	a.setContent(content)

	go func() {
		var (
			cardResults []api.SearchResult
			wikiResults []api.SearchResult
			archResults []string
			pageErr     error
			archErr     error
			wg          sync.WaitGroup
		)

		wg.Add(2)

		// Single wiki search → split into cards + wiki articles
		go func() {
			defer wg.Done()
			cardResults, wikiResults, pageErr = a.svc.SearchAllPages(query)
		}()

		go func() {
			defer wg.Done()
			archResults, archErr = a.svc.SearchArchetypes(query)
		}()

		wg.Wait()

		// Build card result items
		var cardItems []fyne.CanvasObject
		if pageErr != nil {
			cardItems = []fyne.CanvasObject{widget.NewLabel("Search failed: " + pageErr.Error())}
		} else if len(cardResults) == 0 {
			dim := canvas.NewText("No card results.", theme.ColorFGDim)
			dim.TextSize = 13
			cardItems = []fyne.CanvasObject{dim}
		} else {
			for _, r := range cardResults {
				title := r.Title
				btn := widget.NewButton(title, func() {
					a.showCardByName(title)
				})
				btn.Importance = widget.MediumImportance
				btn.Alignment = widget.ButtonAlignLeading
				cardItems = append(cardItems, newTappableButton(btn))
			}
		}

		// Build archetype result items
		var archItems []fyne.CanvasObject
		if archErr != nil {
			archItems = []fyne.CanvasObject{widget.NewLabel("Failed: " + archErr.Error())}
		} else if len(archResults) == 0 {
			dim := canvas.NewText("No archetype results.", theme.ColorFGDim)
			dim.TextSize = 13
			archItems = []fyne.CanvasObject{dim}
		} else {
			for _, name := range archResults {
				archName := name
				btn := widget.NewButton(archName, func() {
					a.showArchetype(archName)
				})
				btn.Importance = widget.LowImportance
				btn.Alignment = widget.ButtonAlignLeading
				archItems = append(archItems, newTappableButton(btn))
			}
		}

		// Build wiki article result items
		var articleItems []fyne.CanvasObject
		if pageErr != nil {
			// Already shown in cards column
		} else if len(wikiResults) == 0 {
			dim := canvas.NewText("No wiki articles.", theme.ColorFGDim)
			dim.TextSize = 13
			articleItems = []fyne.CanvasObject{dim}
		} else {
			for _, r := range wikiResults {
				title := r.Title
				btn := widget.NewButton(title, func() {
					a.showArticle(title)
				})
				btn.Importance = widget.LowImportance
				btn.Alignment = widget.ButtonAlignLeading
				articleItems = append(articleItems, newTappableButton(btn))
			}
		}

		fyne.Do(func() {
			statusLabel.Text = ""
			statusLabel.Refresh()
			for _, item := range cardItems {
				cardResultsList.Add(item)
			}
			cardResultsList.Refresh()
			for _, item := range archItems {
				archetypeResultsList.Add(item)
			}
			archetypeResultsList.Refresh()
			for _, item := range articleItems {
				articleResultsList.Add(item)
			}
			articleResultsList.Refresh()
		})
	}()
}
