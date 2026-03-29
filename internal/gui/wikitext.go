package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/card"
	"github.com/havoczephyr/material-overlay-gui/internal/theme"
)

// wikiTextToRichText converts raw wiki text into a Fyne RichText widget
// with clickable card links that navigate within the app.
func (a *App) wikiTextToRichText(raw string) *widget.RichText {
	if raw == "" {
		return widget.NewRichText()
	}
	cleaned := card.CleanWikiTextPreserveLinks(raw)
	return a.buildRichText(cleaned)
}

// buildRichText builds a RichText widget from pre-cleaned text with [[links]].
func (a *App) buildRichText(cleaned string) *widget.RichText {
	plainText, links := card.ParseWikiLinks(cleaned)

	if len(links) == 0 {
		rt := widget.NewRichText(&widget.TextSegment{
			Text:  plainText,
			Style: widget.RichTextStyleParagraph,
		})
		rt.Wrapping = fyne.TextWrapWord
		return rt
	}

	var segments []widget.RichTextSegment
	cursor := 0

	for _, link := range links {
		if link.Start > cursor {
			segments = append(segments, &widget.TextSegment{
				Text: plainText[cursor:link.Start],
				Style: widget.RichTextStyle{
					Inline: true,
				},
			})
		}

		cardName := link.Target
		segments = append(segments, &widget.HyperlinkSegment{
			Text: link.Display,
			OnTapped: func() {
				a.showCardByName(cardName)
			},
		})

		cursor = link.End
	}

	if cursor < len(plainText) {
		segments = append(segments, &widget.TextSegment{
			Text: plainText[cursor:],
			Style: widget.RichTextStyle{
				Inline: true,
			},
		})
	}

	rt := widget.NewRichText(segments...)
	rt.Wrapping = fyne.TextWrapWord
	return rt
}

// wikiArticleToContent renders a full wiki article with text and tables.
func (a *App) wikiArticleToContent(raw string) fyne.CanvasObject {
	segments := card.ParseArticle(raw)
	var items []fyne.CanvasObject
	for _, seg := range segments {
		if seg.Table != nil {
			items = append(items, container.NewPadded(renderWikiTable(seg.Table)))
		} else if seg.Text != "" {
			items = append(items, a.buildRichText(seg.Text))
		}
	}
	if len(items) == 0 {
		dimText := canvas.NewText("No content available.", theme.ColorFGDim)
		dimText.TextSize = 13
		return dimText
	}
	return container.NewVBox(items...)
}

// renderWikiTable renders a parsed wiki table as a Fyne grid.
func renderWikiTable(table *card.WikiTable) fyne.CanvasObject {
	if len(table.Rows) == 0 {
		return widget.NewLabel("")
	}

	maxCols := 0
	for _, row := range table.Rows {
		if len(row.Cells) > maxCols {
			maxCols = len(row.Cells)
		}
	}
	if maxCols == 0 {
		return widget.NewLabel("")
	}

	grid := container.NewGridWithColumns(maxCols)
	for _, row := range table.Rows {
		for i := 0; i < maxCols; i++ {
			var cellText string
			if i < len(row.Cells) {
				cellText = row.Cells[i]
			}
			if row.IsHeader {
				text := canvas.NewText(cellText, theme.ColorPrimary)
				text.TextStyle.Bold = true
				text.TextSize = 13
				grid.Add(container.NewPadded(text))
			} else {
				text := canvas.NewText(cellText, theme.ColorFG)
				text.TextSize = 13
				grid.Add(container.NewPadded(text))
			}
		}
	}

	bg := canvas.NewRectangle(theme.ColorBGLight)
	bg.CornerRadius = 4
	return container.NewStack(bg, container.NewPadded(grid))
}
