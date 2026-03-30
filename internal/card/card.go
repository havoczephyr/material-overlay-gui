package card

import (
	"html"
	"regexp"
	"strconv"
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
	reFileTag  = regexp.MustCompile(`\[\[(File|Image):([^\]]+)\]\]`)
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
// Strips all markup including links — produces clean text for display in labels.
func StripWikiPageContent(s string) string {
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
	s = rePipeLink.ReplaceAllString(s, "$1")
	s = reLink.ReplaceAllString(s, "$1")
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

// stripTemplates removes all {{...}} template blocks, handling arbitrary nesting.
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

// WikiLink represents a card reference found in wiki text.
type WikiLink struct {
	Start   int    // byte offset in the plain text output
	End     int    // byte offset end in the plain text output
	Target  string // card name to navigate to (the link target)
	Display string // display text shown to the user
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

// ── Shared types for article parsing ────────────────────────────────────

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

// WikiImage represents an image reference extracted from wikitext.
type WikiImage struct {
	Filename string // e.g. "NormalSummon-Diagram.png" (without "File:" prefix)
	Caption  string // last non-keyword pipe segment, stripped of markup
	Width    int    // parsed from "300px", default 300
}

// ContentSegment represents either a text block or an image reference in a parsed article.
type ContentSegment struct {
	Text  string     // non-empty for text segments
	Image *WikiImage // non-nil for image segments
}

// imageKeywords are [[File:...]] parameters that are not captions.
var imageKeywords = map[string]bool{
	"thumb": true, "thumbnail": true, "frame": true, "frameless": true,
	"border": true, "left": true, "right": true, "center": true, "none": true,
	"upright": true, "baseline": true, "middle": true, "sub": true, "super": true,
	"top": true, "text-top": true, "bottom": true, "text-bottom": true,
}

var reSizePx = regexp.MustCompile(`^\d+(x\d+)?px$`)

// ParseArticleWithImages extracts [[File:...]] / [[Image:...]] tags from raw wikitext,
// preserving their positions relative to the text. Returns an interleaved list of
// text segments (plain text via StripWikiPageContent) and image segments.
func ParseArticleWithImages(rawWikitext string) []ContentSegment {
	if rawWikitext == "" {
		return nil
	}

	matches := reFileTag.FindAllStringIndex(rawWikitext, -1)
	if len(matches) == 0 {
		// No images — return a single text segment
		text := StripWikiPageContent(rawWikitext)
		if text == "" {
			return nil
		}
		return []ContentSegment{{Text: text}}
	}

	var segments []ContentSegment
	prev := 0

	for _, loc := range matches {
		// Text before this image tag
		if loc[0] > prev {
			text := StripWikiPageContent(rawWikitext[prev:loc[0]])
			if text != "" {
				segments = append(segments, ContentSegment{Text: text})
			}
		}

		// Parse the image tag
		tag := rawWikitext[loc[0]:loc[1]]
		img := parseFileTag(tag)
		if img != nil {
			segments = append(segments, ContentSegment{Image: img})
		}

		prev = loc[1]
	}

	// Text after the last image tag
	if prev < len(rawWikitext) {
		text := StripWikiPageContent(rawWikitext[prev:])
		if text != "" {
			segments = append(segments, ContentSegment{Text: text})
		}
	}

	return segments
}

// parseFileTag parses a [[File:Name.png|thumb|300px|Caption]] tag into a WikiImage.
func parseFileTag(tag string) *WikiImage {
	// Strip [[ and ]]
	inner := tag[2 : len(tag)-2]

	// Remove "File:" or "Image:" prefix
	if idx := strings.Index(inner, ":"); idx != -1 {
		inner = inner[idx+1:]
	} else {
		return nil
	}

	parts := strings.Split(inner, "|")
	if len(parts) == 0 {
		return nil
	}

	img := &WikiImage{
		Filename: strings.TrimSpace(parts[0]),
		Width:    300,
	}

	// Walk remaining parts: keywords are skipped, size is parsed, last non-keyword is caption
	var captionCandidate string
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		lower := strings.ToLower(p)
		if imageKeywords[lower] {
			continue
		}
		if reSizePx.MatchString(lower) {
			// Parse width from "300px" or "300x200px"
			numStr := strings.TrimSuffix(lower, "px")
			if xIdx := strings.Index(numStr, "x"); xIdx != -1 {
				numStr = numStr[:xIdx]
			}
			if w, err := strconv.Atoi(numStr); err == nil && w > 0 {
				img.Width = w
			}
			continue
		}
		// Anything else is a potential caption (last one wins)
		captionCandidate = p
	}

	if captionCandidate != "" {
		img.Caption = StripWikiMarkup(captionCandidate)
	}

	return img
}

