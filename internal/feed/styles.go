package feed

// DefaultStyleName is the default layout style when none is specified
const DefaultStyleName = "header"

// LayoutStyle defines a post layout format for the TUI.
type LayoutStyle struct {
	// Name is the identifier for the style (e.g., "header")
	Name string
	// DisplayName is the human-readable name (e.g., "Header")
	DisplayName string
}

// AllStyles is the registry of available layout styles.
// Styles cycle in order: header → irc → slack → minimal
var AllStyles = []LayoutStyle{
	{
		Name:        "header",
		DisplayName: "Header",
	},
	{
		Name:        "irc",
		DisplayName: "IRC",
	},
	{
		Name:        "slack",
		DisplayName: "Slack",
	},
	{
		Name:        "minimal",
		DisplayName: "Minimal",
	},
}

// GetStyle returns the style with the given name, or the default style if not found.
func GetStyle(name string) *LayoutStyle {
	for i := range AllStyles {
		if AllStyles[i].Name == name {
			return &AllStyles[i]
		}
	}
	return &AllStyles[0]
}

// NextStyle returns the name of the next style for cycling.
func NextStyle(current string) string {
	for i, s := range AllStyles {
		if s.Name == current {
			return AllStyles[(i+1)%len(AllStyles)].Name
		}
	}
	return AllStyles[0].Name
}
