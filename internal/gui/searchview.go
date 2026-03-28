package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

func (a *App) showSearch(query string) {
	header := sectionHeader("Search Results")
	statusLabel := canvas.NewText("Searching for \""+query+"\"...", theme.ColorFGDim)
	statusLabel.TextSize = 13

	resultsList := container.NewVBox()
	content := container.NewVBox(header, statusLabel, resultsList)
	scrollable := container.NewVScroll(container.NewPadded(content))

	a.setContent(scrollable)

	go func() {
		results, err := a.svc.SearchCards(query)
		if err != nil {
			statusLabel.Text = "Search failed: " + err.Error()
			statusLabel.Refresh()
			return
		}

		if len(results) == 0 {
			statusLabel.Text = "No results found for \"" + query + "\""
			statusLabel.Refresh()
			return
		}

		statusLabel.Text = ""
		statusLabel.Refresh()

		for _, r := range results {
			title := r.Title
			btn := widget.NewButton(title, func() {
				a.showCardByName(title)
			})
			btn.Importance = widget.MediumImportance
			btn.Alignment = widget.ButtonAlignLeading
			resultsList.Add(container.NewGridWrap(fyne.NewSize(500, 40), btn))
		}
		resultsList.Refresh()
	}()
}
