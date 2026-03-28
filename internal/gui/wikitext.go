package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/havoczephyr/material-overlay-gui/internal/card"
)

// wikiTextToRichText converts raw wiki text into a Fyne RichText widget
// with clickable card links that navigate within the app.
func (a *App) wikiTextToRichText(raw string) *widget.RichText {
	if raw == "" {
		return widget.NewRichText()
	}

	// Clean the text but keep [[links]] intact
	cleaned := card.CleanWikiTextPreserveLinks(raw)

	// Parse out links and get plain text with positions
	plainText, links := card.ParseWikiLinks(cleaned)

	if len(links) == 0 {
		rt := widget.NewRichText(&widget.TextSegment{
			Text:  plainText,
			Style: widget.RichTextStyleParagraph,
		})
		rt.Wrapping = fyne.TextWrapWord
		return rt
	}

	// Build segments: alternate between text and hyperlinks
	var segments []widget.RichTextSegment
	cursor := 0

	for _, link := range links {
		// Text before this link
		if link.Start > cursor {
			segments = append(segments, &widget.TextSegment{
				Text: plainText[cursor:link.Start],
				Style: widget.RichTextStyle{
					Inline: true,
				},
			})
		}

		// The clickable link
		cardName := link.Target
		segments = append(segments, &widget.HyperlinkSegment{
			Text: link.Display,
			OnTapped: func() {
				a.showCardByName(cardName)
			},
		})

		cursor = link.End
	}

	// Remaining text after last link
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
