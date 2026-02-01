package feed

// DefaultLayoutName is the default layout when none is specified
const DefaultLayoutName = "comfy"

// LayoutStyle defines a post layout format for the TUI.
type LayoutStyle struct {
	// Name is the identifier for the layout (e.g., "comfy")
	Name string
	// DisplayName is the human-readable name (e.g., "Comfy")
	DisplayName string
}

// AllLayouts is the registry of available layouts.
// Layouts cycle in order: dense → comfy → relaxed
var AllLayouts = []LayoutStyle{
	{
		Name:        "dense",
		DisplayName: "Dense",
	},
	{
		Name:        "comfy",
		DisplayName: "Comfy",
	},
	{
		Name:        "relaxed",
		DisplayName: "Relaxed",
	},
}

// GetLayout returns the layout with the given name, or the default layout if not found.
func GetLayout(name string) *LayoutStyle {
	for i := range AllLayouts {
		if AllLayouts[i].Name == name {
			return &AllLayouts[i]
		}
	}
	// Return default (comfy)
	for i := range AllLayouts {
		if AllLayouts[i].Name == DefaultLayoutName {
			return &AllLayouts[i]
		}
	}
	return &AllLayouts[0]
}

// NextLayout returns the name of the next layout for cycling.
func NextLayout(current string) string {
	for i, l := range AllLayouts {
		if l.Name == current {
			return AllLayouts[(i+1)%len(AllLayouts)].Name
		}
	}
	return AllLayouts[0].Name
}

// PrevLayout returns the name of the previous layout for reverse cycling.
func PrevLayout(current string) string {
	for i, l := range AllLayouts {
		if l.Name == current {
			return AllLayouts[(i-1+len(AllLayouts))%len(AllLayouts)].Name
		}
	}
	return AllLayouts[0].Name
}

// Backward compatibility aliases
var AllStyles = AllLayouts

func GetStyle(name string) *LayoutStyle {
	return GetLayout(name)
}

func NextStyle(current string) string {
	return NextLayout(current)
}
