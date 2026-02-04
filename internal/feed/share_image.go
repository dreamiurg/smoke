// Package feed provides image rendering for post sharing.
package feed

import (
	"bytes"
	"image/color"
	"image/png"

	"github.com/fogleman/gg"
)

// ImageDimensions defines standard social media image sizes
type ImageDimensions struct {
	Width  int
	Height int
	Name   string
}

var (
	// SquareImage is 1200x1200 for Instagram, etc.
	SquareImage = ImageDimensions{Width: 1200, Height: 1200, Name: "square"}
	// LandscapeImage is 1200x630 for Twitter/OG cards
	LandscapeImage = ImageDimensions{Width: 1200, Height: 630, Name: "landscape"}
)

// RenderShareCard renders a post as a shareable PNG image.
// Uses theme colors for Carbon-style terminal aesthetic.
func RenderShareCard(post *Post, theme *Theme, dims ImageDimensions) ([]byte, error) {
	dc := gg.NewContext(dims.Width, dims.Height)

	// Convert theme colors to Go color.Color (use Dark variant for terminal aesthetic)
	bgColor := hexToColor(theme.Background.Dark)
	textColor := hexToColor(theme.Text.Dark)
	accentColor := hexToColor(theme.Accent.Dark)

	// Fill background
	dc.SetColor(bgColor)
	dc.Clear()

	// Draw rounded rectangle card (with some padding)
	padding := float64(dims.Width) * 0.05
	cardWidth := float64(dims.Width) - padding*2
	cardHeight := float64(dims.Height) - padding*2
	cornerRadius := 20.0

	dc.SetColor(hexToColor(theme.BackgroundSecondary.Dark))
	drawRoundedRect(dc, padding, padding, cardWidth, cardHeight, cornerRadius)
	dc.Fill()

	// Card inner padding
	innerPadding := padding + 40

	// Draw window controls (Carbon-style dots)
	dotY := innerPadding + 10
	dotRadius := 7.0
	dotSpacing := 20.0

	colors := []color.Color{
		color.RGBA{255, 95, 86, 255},  // Red
		color.RGBA{255, 189, 46, 255}, // Yellow
		color.RGBA{39, 201, 63, 255},  // Green
	}
	for i, c := range colors {
		dc.SetColor(c)
		dc.DrawCircle(innerPadding+float64(i)*dotSpacing+10, dotY, dotRadius)
		dc.Fill()
	}

	// Draw handle
	handleY := dotY + 50
	fontSize := float64(dims.Width) * 0.025
	if err := dc.LoadFontFace("/System/Library/Fonts/SFNSMono.ttf", fontSize); err != nil {
		// Fallback to default if font not found
		_ = dc.LoadFontFace("/Library/Fonts/Courier New.ttf", fontSize)
	}
	handle := post.Author
	if handle == "" {
		handle = "anonymous"
	}

	agent, project := SplitIdentity(handle)
	agentColor := agentColorForTheme(agent, theme)
	projectColor := hexToColor(theme.TextMuted.Dark)

	dc.SetColor(agentColor)
	dc.DrawString(agent, innerPadding, handleY)

	handleWidth := 0.0
	agentWidth, _ := dc.MeasureString(agent)
	handleWidth += agentWidth

	if project != "" {
		dc.SetColor(projectColor)
		dc.DrawString("@"+project, innerPadding+agentWidth, handleY)
		projectWidth, _ := dc.MeasureString("@" + project)
		handleWidth += projectWidth
	}

	caller := ResolveCallerTag(post)
	if caller != "" {
		dc.SetColor(projectColor)
		dc.DrawString("  via "+caller, innerPadding+handleWidth, handleY)
	}

	// Draw content
	contentY := handleY + fontSize*2
	dc.SetColor(textColor)
	contentFontSize := fontSize * 1.5
	_ = dc.LoadFontFace("/System/Library/Fonts/SFNSMono.ttf", contentFontSize)

	// Word wrap the content
	maxWidth := cardWidth - 80
	lines := dc.WordWrap(post.Content, maxWidth)
	lineHeight := contentFontSize * 1.4
	for i, line := range lines {
		if contentY+float64(i)*lineHeight > float64(dims.Height)-innerPadding-60 {
			break // Stop if we run out of space
		}
		dc.DrawString(line, innerPadding, contentY+float64(i)*lineHeight)
	}

	// Draw footer
	footerY := float64(dims.Height) - innerPadding
	dc.SetColor(accentColor)
	footerFontSize := fontSize * 0.8
	_ = dc.LoadFontFace("/System/Library/Fonts/SFNSMono.ttf", footerFontSize)
	dc.DrawString(ShareFooter, innerPadding, footerY)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// agentColorForTheme returns the agent name color based on theme palette.
func agentColorForTheme(agent string, theme *Theme) color.Color {
	if theme == nil || len(theme.AgentColors) == 0 {
		return color.Black
	}
	idx := hashString(agent) % len(theme.AgentColors)
	return hexToColor(string(theme.AgentColors[idx]))
}

// hexToColor converts a hex color string to color.Color
func hexToColor(hex string) color.Color {
	if len(hex) == 0 {
		return color.Black
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return color.Black
	}

	var r, g, b uint8
	n, _ := parseHex(hex)
	r = uint8(n >> 16)
	g = uint8(n >> 8)
	b = uint8(n)

	return color.RGBA{r, g, b, 255}
}

// parseHex parses a hex string to an integer
func parseHex(s string) (int64, error) {
	var result int64
	for _, c := range s {
		result *= 16
		switch {
		case c >= '0' && c <= '9':
			result += int64(c - '0')
		case c >= 'a' && c <= 'f':
			result += int64(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			result += int64(c - 'A' + 10)
		}
	}
	return result, nil
}

// drawRoundedRect draws a rounded rectangle path
func drawRoundedRect(dc *gg.Context, x, y, w, h, r float64) {
	dc.MoveTo(x+r, y)
	dc.LineTo(x+w-r, y)
	dc.QuadraticTo(x+w, y, x+w, y+r)
	dc.LineTo(x+w, y+h-r)
	dc.QuadraticTo(x+w, y+h, x+w-r, y+h)
	dc.LineTo(x+r, y+h)
	dc.QuadraticTo(x, y+h, x, y+h-r)
	dc.LineTo(x, y+r)
	dc.QuadraticTo(x, y, x+r, y)
	dc.ClosePath()
}
