package theme

import (
	"image/color"
	"strconv"
)

// Color palette — Yu-Gi-Oh themed with material(Overlay) branding.
var (
	// Brand colors
	ColorGold      = parseHex("#FFD700")
	ColorDarkGold  = parseHex("#B8860B")
	ColorWhite     = parseHex("#ffffff")
	ColorBG        = parseHex("#1c1c1c")
	ColorBGLight   = parseHex("#262626")
	ColorFG        = parseHex("#e0e0e0")
	ColorFGDim     = parseHex("#888888")
	ColorAccent    = parseHex("#444444")
	ColorBrand     = parseHex("#e94560")

	// Decorative
	ColorDecoBlack  = parseHex("#111111")
	ColorDecoBright = parseHex("#00e5ff")
	ColorDecoDark   = parseHex("#00acc1")

	// Attribute colors
	ColorDark   = parseHex("#9b59b6")
	ColorLight  = parseHex("#f1c40f")
	ColorFire   = parseHex("#e74c3c")
	ColorWater  = parseHex("#3498db")
	ColorEarth  = parseHex("#8B4513")
	ColorWind   = parseHex("#2ecc71")
	ColorDivine = parseHex("#FFD700")

	// Status badge colors
	ColorUnlimited = parseHex("#2ecc71")
	ColorLimited   = parseHex("#f39c12")
	ColorSemiLtd   = parseHex("#e67e22")
	ColorForbidden = parseHex("#e74c3c")

	// ATK/DEF
	ColorATK = parseHex("#e74c3c")
	ColorDEF = parseHex("#3498db")
)

// AttributeColor returns the color for a Yu-Gi-Oh card attribute.
func AttributeColor(attr string) color.Color {
	switch attr {
	case "DARK":
		return ColorDark
	case "LIGHT":
		return ColorLight
	case "FIRE":
		return ColorFire
	case "WATER":
		return ColorWater
	case "EARTH":
		return ColorEarth
	case "WIND":
		return ColorWind
	case "DIVINE":
		return ColorDivine
	default:
		return ColorFGDim
	}
}

// StatusColor returns the color for a card's legal status.
func StatusColor(status string) color.Color {
	switch status {
	case "Unlimited":
		return ColorUnlimited
	case "Limited":
		return ColorLimited
	case "Semi-Limited":
		return ColorSemiLtd
	case "Forbidden":
		return ColorForbidden
	default:
		return ColorFGDim
	}
}

// GenesysCostColor returns a color for a Genesys point cost string.
func GenesysCostColor(cost string) color.Color {
	if cost == "Banned" {
		return ColorForbidden
	}
	points, err := strconv.Atoi(cost)
	if err != nil {
		return ColorFGDim
	}
	switch {
	case points == 0:
		return ColorUnlimited
	case points <= 34:
		return parseHex("#27ae60")
	case points <= 74:
		return ColorLimited
	default:
		return ColorForbidden
	}
}

func parseHex(hex string) color.NRGBA {
	if hex[0] == '#' {
		hex = hex[1:]
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xFF}
}
