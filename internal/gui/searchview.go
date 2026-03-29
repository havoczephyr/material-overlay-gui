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

	// Two result columns
	cardResultsList := container.NewVBox()
	archetypeResultsList := container.NewVBox()

	cardColumn := container.NewVBox(sectionHeader("Cards"), cardResultsList)
	archColumn := container.NewVBox(sectionHeader("Archetypes"), archetypeResultsList)

	resultsSplit := container.NewHSplit(
		container.NewVScroll(container.NewPadded(cardColumn)),
		container.NewVScroll(container.NewPadded(archColumn)),
	)
	resultsSplit.SetOffset(0.65)

	content := container.NewBorder(
		container.NewPadded(container.NewVBox(header, statusLabel)),
		nil, nil, nil,
		resultsSplit,
	)
	a.setContent(content)

	go func() {
		var (
			cardResults []api.SearchResult
			archResults []string
			cardErr     error
			archErr     error
			wg          sync.WaitGroup
		)

		wg.Add(2)

		go func() {
			defer wg.Done()
			cardResults, cardErr = a.svc.SearchCards(query)
		}()

		go func() {
			defer wg.Done()
			archResults, archErr = a.svc.SearchArchetypes(query)
		}()

		wg.Wait()

		// Build card result items
		var cardItems []fyne.CanvasObject
		if cardErr != nil {
			cardItems = []fyne.CanvasObject{widget.NewLabel("Search failed: " + cardErr.Error())}
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
		})
	}()
}
