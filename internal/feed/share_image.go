// Package feed provides image rendering for post sharing.
package feed

import (
	"bytes"
	"image/color"
	"image/png"
	"math"
	"strings"

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

// renderCardBackground draws the card background and window controls.
func renderCardBackground(dc *gg.Context, theme *Theme, dims ImageDimensions) (padding, innerPadding, cardWidth float64) {
	bgColor := hexToColor(theme.Background.Dark)
	dc.SetColor(bgColor)
	dc.Clear()

	padding = float64(dims.Width) * 0.05
	cardWidth = float64(dims.Width) - padding*2
	cardHeight := float64(dims.Height) - padding*2
	cornerRadius := 20.0

	dc.SetColor(hexToColor(theme.BackgroundSecondary.Dark))
	drawRoundedRect(dc, roundedRect{padding, padding, cardWidth, cardHeight, cornerRadius})
	dc.Fill()

	innerPadding = padding + 40

	dotY := innerPadding + 10
	dotRadius := 7.0
	dotSpacing := 20.0
	colors := []color.Color{
		color.RGBA{255, 95, 86, 255},
		color.RGBA{255, 189, 46, 255},
		color.RGBA{39, 201, 63, 255},
	}
	for i, c := range colors {
		dc.SetColor(c)
		dc.DrawCircle(innerPadding+float64(i)*dotSpacing+10, dotY, dotRadius)
		dc.Fill()
	}

	return padding, innerPadding, cardWidth
}

// cardLayout bundles layout parameters for card handle rendering.
type cardLayout struct {
	innerPadding float64
	dotY         float64
	fontSize     float64
}

// renderCardHandle draws the author handle (agent@project [caller]).
// Returns the Y position below the handle for content placement.
func renderCardHandle(dc *gg.Context, post *Post, theme *Theme, layout cardLayout) float64 {
	handleY := layout.dotY + 50
	loadMonoFont(dc, layout.fontSize)
	handle := post.Author
	if handle == "" {
		handle = "anonymous"
	}

	agent, project := SplitIdentity(handle)
	projectColor := hexToColor(theme.TextMuted.Dark)

	dc.SetColor(agentColorForTheme(agent, theme))
	dc.DrawString(agent, layout.innerPadding, handleY)

	agentWidth, _ := dc.MeasureString(agent)
	handleWidth := agentWidth

	if project != "" {
		dc.SetColor(projectColor)
		dc.DrawString("@"+project, layout.innerPadding+agentWidth, handleY)
		projectWidth, _ := dc.MeasureString("@" + project)
		handleWidth += projectWidth
	}

	if caller := ResolveCallerTag(post); caller != "" {
		dc.SetColor(projectColor)
		dc.DrawString(" ("+caller+")", layout.innerPadding+handleWidth, handleY)
	}

	return handleY
}

// contentLayout bundles layout parameters for card content rendering.
type contentLayout struct {
	innerPadding    float64
	contentY        float64
	cardWidth       float64
	availableHeight float64
	fontSize        float64
}

// renderCardContent draws the post content with auto-sizing font.
func renderCardContent(dc *gg.Context, post *Post, theme *Theme, cl contentLayout) {
	dc.SetColor(hexToColor(theme.Text.Dark))
	contentFontSize := cl.fontSize * 1.5
	minFontSize := cl.fontSize * 0.8
	maxWidth := cl.cardWidth - 80

	var lines []string
	var lineHeight float64
	for {
		loadMonoFont(dc, contentFontSize)
		lines = dc.WordWrap(post.Content, maxWidth)
		lineHeight = contentFontSize * 1.4
		if lineHeight*float64(len(lines)) <= cl.availableHeight || contentFontSize <= minFontSize {
			break
		}
		contentFontSize -= 1.0
	}

	if cl.availableHeight <= 0 || len(lines) == 0 {
		return
	}

	maxLines := int(math.Floor(cl.availableHeight / lineHeight))
	if maxLines <= 0 {
		return
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[maxLines-1] = strings.TrimRight(lines[maxLines-1], " ") + "…"
	}
	for i, line := range lines {
		dc.DrawString(line, cl.innerPadding, cl.contentY+float64(i)*lineHeight)
	}
}

// RenderShareCard renders a post as a shareable PNG image.
// Uses theme colors for Carbon-style terminal aesthetic.
func RenderShareCard(post *Post, theme *Theme, dims ImageDimensions) ([]byte, error) {
	dc := gg.NewContext(dims.Width, dims.Height)

	_, innerPadding, cardWidth := renderCardBackground(dc, theme, dims)

	dotY := innerPadding + 10
	fontSize := float64(dims.Width) * 0.025
	handleY := renderCardHandle(dc, post, theme, cardLayout{innerPadding, dotY, fontSize})

	contentY := handleY + fontSize*2
	footerFontSize := fontSize * 0.8
	footerY := float64(dims.Height) - innerPadding
	contentMaxY := footerY - (footerFontSize * 1.6)
	if contentMaxY < contentY {
		contentMaxY = contentY
	}

	renderCardContent(dc, post, theme, contentLayout{
		innerPadding:    innerPadding,
		contentY:        contentY,
		cardWidth:       cardWidth,
		availableHeight: contentMaxY - contentY,
		fontSize:        fontSize,
	})

	dc.SetColor(hexToColor(theme.Accent.Dark))
	loadMonoFont(dc, footerFontSize)
	dc.DrawString(ShareFooter, innerPadding, footerY)

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

// roundedRect bundles the dimensions for a rounded rectangle.
type roundedRect struct {
	x, y, w, h, r float64
}

// drawRoundedRect draws a rounded rectangle path
func drawRoundedRect(dc *gg.Context, rect roundedRect) {
	dc.MoveTo(rect.x+rect.r, rect.y)
	dc.LineTo(rect.x+rect.w-rect.r, rect.y)
	dc.QuadraticTo(rect.x+rect.w, rect.y, rect.x+rect.w, rect.y+rect.r)
	dc.LineTo(rect.x+rect.w, rect.y+rect.h-rect.r)
	dc.QuadraticTo(rect.x+rect.w, rect.y+rect.h, rect.x+rect.w-rect.r, rect.y+rect.h)
	dc.LineTo(rect.x+rect.r, rect.y+rect.h)
	dc.QuadraticTo(rect.x, rect.y+rect.h, rect.x, rect.y+rect.h-rect.r)
	dc.LineTo(rect.x, rect.y+rect.r)
	dc.QuadraticTo(rect.x, rect.y, rect.x+rect.r, rect.y)
	dc.ClosePath()
}

func loadMonoFont(dc *gg.Context, size float64) {
	if err := dc.LoadFontFace("/System/Library/Fonts/SFNSMono.ttf", size); err != nil {
		_ = dc.LoadFontFace("/Library/Fonts/Courier New.ttf", size)
	}
}
