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
	// wiki table cell attributes (e.g., colspan="2" | content)
	reCellAttr = regexp.MustCompile(`^(?:\s*[\w-]+\s*=\s*"[^"]*"\s*)+\|`)
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

	// Remove templates (depth-tracking for arbitrary nesting)
	s = stripTemplates(s)

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

// WikiLink represents a card reference found in wiki text.
type WikiLink struct {
	Start   int    // byte offset in the plain text output
	End     int    // byte offset end in the plain text output
	Target  string // card name to navigate to (the link target)
	Display string // display text shown to the user
}

// CleanWikiTextPreserveLinks applies the same cleanup as StripWikiPageContent
// but keeps [[...]] link markup intact for later parsing.
func CleanWikiTextPreserveLinks(s string) string {
	if s == "" {
		return ""
	}

	s = reRefBlock.ReplaceAllString(s, "")
	s = reRefSelf.ReplaceAllString(s, "")

	s = stripTemplates(s)

	s = reTableStart.ReplaceAllString(s, "")
	s = reTableEnd.ReplaceAllString(s, "")
	s = reTableRowSep.ReplaceAllString(s, "")

	s = reSectionHeader.ReplaceAllString(s, "$1")
	s = reBold.ReplaceAllString(s, "$1")
	s = reItalic.ReplaceAllString(s, "$1")
	s = reBR.ReplaceAllString(s, "\n")
	s = reHTML.ReplaceAllString(s, "")
	s = html.UnescapeString(s)

	// Process lines but preserve [[...]] links
	var result []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

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

		if strings.HasPrefix(trimmed, "!") {
			trimmed = strings.TrimSpace(strings.TrimLeft(trimmed, "!"))
			if trimmed == "" {
				continue
			}
		}

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

// ParseWikiLinks extracts [[...]] links from cleaned wiki text.
// Returns the plain text with links replaced by display text, and the link positions.
func ParseWikiLinks(cleaned string) (string, []WikiLink) {
	var links []WikiLink
	var out strings.Builder
	offset := 0

	// Process pipe links first: [[target|display]]
	// Then simple links: [[target]]
	// We need to handle both in a single pass.
	remaining := cleaned
	for len(remaining) > 0 {
		// Find next [[ marker
		start := strings.Index(remaining, "[[")
		if start == -1 {
			out.WriteString(remaining)
			break
		}

		// Write text before the link
		out.WriteString(remaining[:start])
		offset += start

		// Find closing ]]
		end := strings.Index(remaining[start:], "]]")
		if end == -1 {
			// No closing brackets - write as-is
			out.WriteString(remaining[start:])
			break
		}
		end += start // adjust to absolute position

		inner := remaining[start+2 : end]
		remaining = remaining[end+2:]

		// Parse link: [[target|display]] or [[target]]
		var target, display string
		if pipeIdx := strings.Index(inner, "|"); pipeIdx != -1 {
			target = strings.TrimSpace(inner[:pipeIdx])
			display = strings.TrimSpace(inner[pipeIdx+1:])
		} else {
			target = strings.TrimSpace(inner)
			display = target
		}

		// Strip anchor fragments (e.g., "Card Tips:Dark Magician#English" → "Card Tips:Dark Magician")
		if hashIdx := strings.Index(target, "#"); hashIdx != -1 {
			target = target[:hashIdx]
		}

		// Strip namespace prefixes for card navigation (e.g., "Card Tips:Dark Magician" → keep as-is for non-card pages)
		linkStart := offset
		linkEnd := offset + len(display)

		links = append(links, WikiLink{
			Start:   linkStart,
			End:     linkEnd,
			Target:  target,
			Display: display,
		})

		out.WriteString(display)
		offset += len(display)
	}

	return out.String(), links
}

// stripTemplates removes all {{...}} template blocks, properly handling nesting.
func stripTemplates(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '{' && s[i+1] == '{' {
			depth := 1
			i += 2
			for i < len(s) && depth > 0 {
				if i+1 < len(s) && s[i] == '{' && s[i+1] == '{' {
					depth++
					i += 2
				} else if i+1 < len(s) && s[i] == '}' && s[i+1] == '}' {
					depth--
					i += 2
				} else {
					i++
				}
			}
		} else {
			out.WriteByte(s[i])
			i++
		}
	}
	return out.String()
}

// ── Article parsing (tables + text segments) ────────────────────────────

// WikiTable represents a parsed wiki table.
type WikiTable struct {
	Rows []WikiTableRow
}

// WikiTableRow represents a single row in a wiki table.
type WikiTableRow struct {
	Cells    []string
	IsHeader bool
}

// ArticleSegment represents a part of a wiki article: either text or a table.
type ArticleSegment struct {
	Text  string     // non-empty for text segments (with [[links]] preserved)
	Table *WikiTable // non-nil for table segments
}

// ParseArticle extracts structured content from raw wiki article text.
// Returns interleaved text and table segments.
func ParseArticle(raw string) []ArticleSegment {
	if raw == "" {
		return nil
	}

	s := reRefBlock.ReplaceAllString(raw, "")
	s = reRefSelf.ReplaceAllString(s, "")
	s = stripTemplates(s)

	return extractArticleSegments(s)
}

func extractArticleSegments(s string) []ArticleSegment {
	var segments []ArticleSegment

	for {
		start := strings.Index(s, "{|")
		if start == -1 {
			cleaned := cleanArticleText(s)
			if cleaned != "" {
				segments = append(segments, ArticleSegment{Text: cleaned})
			}
			break
		}

		before := s[:start]
		cleaned := cleanArticleText(before)
		if cleaned != "" {
			segments = append(segments, ArticleSegment{Text: cleaned})
		}

		inner := s[start+2:]
		end := findTableEnd(inner)
		if end == -1 {
			cleaned = cleanArticleText(s[start:])
			if cleaned != "" {
				segments = append(segments, ArticleSegment{Text: cleaned})
			}
			break
		}

		tableContent := inner[:end]
		table := parseWikiTable(tableContent)
		if len(table.Rows) > 0 {
			segments = append(segments, ArticleSegment{Table: &table})
		}

		s = inner[end+2:]
	}

	return segments
}

// findTableEnd finds the matching |} for a {| table start.
// s should be the content after the opening {|.
func findTableEnd(s string) int {
	depth := 1
	i := 0
	for i < len(s) {
		if i+1 < len(s) {
			if s[i] == '{' && s[i+1] == '|' {
				depth++
				i += 2
				continue
			}
			if s[i] == '|' && s[i+1] == '}' {
				depth--
				if depth == 0 {
					return i
				}
				i += 2
				continue
			}
		}
		i++
	}
	return -1
}

// cleanArticleText cleans wiki text for display, preserving [[links]].
func cleanArticleText(s string) string {
	s = reTableStart.ReplaceAllString(s, "")
	s = reTableEnd.ReplaceAllString(s, "")
	s = reTableRowSep.ReplaceAllString(s, "")
	s = reSectionHeader.ReplaceAllString(s, "$1")
	s = reBold.ReplaceAllString(s, "$1")
	s = reItalic.ReplaceAllString(s, "$1")
	s = reBR.ReplaceAllString(s, "\n")
	s = reHTML.ReplaceAllString(s, "")
	s = html.UnescapeString(s)

	var result []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
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
		result = append(result, trimmed)
	}
	return strings.Join(result, "\n")
}

func parseWikiTable(content string) WikiTable {
	var table WikiTable
	lines := strings.Split(content, "\n")
	var currentRow WikiTableRow
	inRow := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "|-") {
			if inRow && len(currentRow.Cells) > 0 {
				table.Rows = append(table.Rows, currentRow)
			}
			currentRow = WikiTableRow{}
			inRow = true
			continue
		}

		if strings.HasPrefix(trimmed, "!") {
			if !inRow {
				inRow = true
			}
			currentRow.IsHeader = true
			cellContent := strings.TrimSpace(trimmed[1:])
			cells := strings.Split(cellContent, "!!")
			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				cell = stripCellAttrs(cell)
				cell = StripWikiMarkup(cell)
				cell = strings.TrimSpace(cell)
				if cell != "" {
					currentRow.Cells = append(currentRow.Cells, cell)
				}
			}
			continue
		}

		if strings.HasPrefix(trimmed, "|") {
			if !inRow {
				inRow = true
			}
			cellContent := strings.TrimSpace(trimmed[1:])
			cells := strings.Split(cellContent, "||")
			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				cell = stripCellAttrs(cell)
				cell = StripWikiMarkup(cell)
				cell = strings.TrimSpace(cell)
				if cell != "" {
					currentRow.Cells = append(currentRow.Cells, cell)
				}
			}
			continue
		}
	}

	if inRow && len(currentRow.Cells) > 0 {
		table.Rows = append(table.Rows, currentRow)
	}

	return table
}

func stripCellAttrs(cell string) string {
	loc := reCellAttr.FindStringIndex(cell)
	if loc != nil {
		return strings.TrimSpace(cell[loc[1]:])
	}
	return cell
}
