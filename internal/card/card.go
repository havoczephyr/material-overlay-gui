package card

import (
	"html"
	"regexp"
	"strings"

	"github.com/havoczephyr/material-overlay-gui/internal/api"
)

type Card struct {
	Name          string   `json:"name"`
	CardType      string   `json:"card_type"`
	Attribute     string   `json:"attribute,omitempty"`
	ATK           string   `json:"atk,omitempty"`
	DEF           string   `json:"def,omitempty"`
	Level         string   `json:"level,omitempty"`
	Rank          string   `json:"rank,omitempty"`
	PendulumScale string   `json:"pendulum_scale,omitempty"`
	LinkArrows    []string `json:"link_arrows,omitempty"`
	Types         string   `json:"types,omitempty"`
	Lore          string   `json:"lore"`
	Property      string   `json:"property,omitempty"`
	Archseries    []string `json:"archseries,omitempty"`
	Password      string   `json:"password,omitempty"`
	TCGStatus     string   `json:"tcg_status,omitempty"`
	OCGStatus     string   `json:"ocg_status,omitempty"`
	GenesysCost   string   `json:"genesys_cost,omitempty"`
	WikiURL       string   `json:"wiki_url"`
	ImageURL      string   `json:"image_url,omitempty"`
	CardSets      []CardSetEntry `json:"card_sets,omitempty"`
	Tips          string   `json:"tips,omitempty"`
	Trivia        string   `json:"trivia,omitempty"`
}

type CardSetEntry struct {
	SetName string `json:"set_name"`
	SetCode string `json:"set_code"`
	Rarity  string `json:"rarity"`
}

// Page-type properties (return []SMWPageValue, extract .Fulltext).
var pageProps = map[string]bool{
	"Card type":  true,
	"Attribute":  true,
	"Archseries": true,
	"TCG status": true,
	"OCG status": true,
}

func NewCardFromSMW(name string, entry api.SMWResultEntry) Card {
	c := Card{
		Name:    entry.Fulltext,
		WikiURL: entry.Fullurl,
	}

	for prop, raw := range entry.Printouts {
		var values []string
		if pageProps[prop] {
			for _, pv := range api.ExtractPageValues(raw) {
				values = append(values, pv.Fulltext)
			}
		} else {
			values = api.ExtractTextValues(raw)
		}

		if len(values) == 0 {
			continue
		}

		switch prop {
		case "Card type":
			c.CardType = values[0]
		case "Attribute":
			c.Attribute = values[0]
		case "ATK string":
			c.ATK = values[0]
		case "DEF string":
			c.DEF = values[0]
		case "Level string":
			c.Level = values[0]
		case "Rank string":
			c.Rank = values[0]
		case "Pendulum Scale string":
			c.PendulumScale = values[0]
		case "Link Arrows":
			c.LinkArrows = values
		case "Types":
			c.Types = StripWikiMarkup(values[0])
		case "Lore":
			c.Lore = StripWikiMarkup(values[0])
		case "Property":
			c.Property = values[0]
		case "Archseries":
			c.Archseries = values
		case "Password":
			c.Password = values[0]
		case "TCG status":
			c.TCGStatus = values[0]
		case "OCG status":
			c.OCGStatus = values[0]
		}
	}

	return c
}

func (c *Card) IsMonster() bool {
	return c.CardType == "Monster Card"
}

func (c *Card) IsLink() bool {
	return len(c.LinkArrows) > 0
}

func (c *Card) IsPendulum() bool {
	return c.PendulumScale != ""
}

var (
	// [[display|text]] -> text
	rePipeLink = regexp.MustCompile(`\[\[[^\]]*\|([^\]]*)\]\]`)
	// [[link]] -> link
	reLink = regexp.MustCompile(`\[\[([^\]]*)\]\]`)
	// ''italic'' -> italic
	reItalic = regexp.MustCompile(`''([^']*?)''`)
	// '''bold''' -> bold (must run before italic)
	reBold = regexp.MustCompile(`'''(.*?)'''`)
	// <br />, <br/>, <br> -> newline
	reBR = regexp.MustCompile(`<br\s*/?>`)
	// <ref>...</ref> (inline references)
	reRefBlock = regexp.MustCompile(`(?s)<ref[^>]*>.*?</ref>`)
	// <ref .../> (self-closing references)
	reRefSelf = regexp.MustCompile(`<ref[^/]*/\s*>`)
	// {{template}} (non-nested — run multiple passes for nesting)
	reTemplate = regexp.MustCompile(`\{\{[^{}]*\}\}`)
	// == Section Header == -> Section Header
	reSectionHeader = regexp.MustCompile(`={2,}\s*(.*?)\s*={2,}`)
	// any remaining HTML tags
	reHTML = regexp.MustCompile(`<[^>]+>`)
	// wiki table markup
	reTableStart  = regexp.MustCompile(`(?m)^\{\|.*$`)
	reTableEnd    = regexp.MustCompile(`(?m)^\|\}.*$`)
	reTableRowSep = regexp.MustCompile(`(?m)^\|-.*$`)
)

func StripWikiMarkup(s string) string {
	s = rePipeLink.ReplaceAllString(s, "$1")
	s = reLink.ReplaceAllString(s, "$1")
	s = reItalic.ReplaceAllString(s, "$1")
	s = reBR.ReplaceAllString(s, "\n")
	s = reHTML.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.TrimSpace(s)
	return s
}

// StripWikiPageContent converts raw wikitext from a wiki page into readable plain text.
// Handles bullets, templates, references, table markup, links, and formatting.
func StripWikiPageContent(s string) string {
	if s == "" {
		return ""
	}

	// Remove references
	s = reRefBlock.ReplaceAllString(s, "")
	s = reRefSelf.ReplaceAllString(s, "")

	// Remove templates (multiple passes for nesting)
	for i := 0; i < 5; i++ {
		prev := s
		s = reTemplate.ReplaceAllString(s, "")
		if s == prev {
			break
		}
	}

	// Remove table markup
	s = reTableStart.ReplaceAllString(s, "")
	s = reTableEnd.ReplaceAllString(s, "")
	s = reTableRowSep.ReplaceAllString(s, "")

	// Strip section headers
	s = reSectionHeader.ReplaceAllString(s, "$1")

	// Strip bold before italic (''' before '')
	s = reBold.ReplaceAllString(s, "$1")

	// Apply standard wiki markup stripping
	s = rePipeLink.ReplaceAllString(s, "$1")
	s = reLink.ReplaceAllString(s, "$1")
	s = reItalic.ReplaceAllString(s, "$1")
	s = reBR.ReplaceAllString(s, "\n")
	s = reHTML.ReplaceAllString(s, "")
	s = html.UnescapeString(s)

	// Process lines: convert bullets, strip table cell prefixes, drop empties
	var result []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Wiki bullet points (* or **) -> bullet character
		if strings.HasPrefix(trimmed, "*") {
			depth := 0
			for _, ch := range trimmed {
				if ch == '*' {
					depth++
				} else {
					break
				}
			}
			prefix := "• "
			if depth > 1 {
				prefix = "  • "
			}
			trimmed = prefix + strings.TrimSpace(trimmed[depth:])
		}

		// Table header cells (! or !!)
		if strings.HasPrefix(trimmed, "!") {
			trimmed = strings.TrimSpace(strings.TrimLeft(trimmed, "!"))
			if trimmed == "" {
				continue
			}
		}

		// Table data cells (| prefix)
		if strings.HasPrefix(trimmed, "|") {
			trimmed = strings.TrimSpace(trimmed[1:])
			if trimmed == "" {
				continue
			}
		}

		result = append(result, trimmed)
	}

	return strings.Join(result, "\n")
}
