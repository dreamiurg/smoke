package feed

// ContrastLevel defines styling rules for identity display.
type ContrastLevel struct {
	// Name is the identifier for the contrast level (e.g., "high", "medium", "low")
	Name string
	// DisplayName is the human-readable name (e.g., "High", "Medium", "Low")
	DisplayName string
	// AgentBold indicates whether agent name is displayed bold
	AgentBold bool
	// AgentColored indicates whether agent uses theme color
	AgentColored bool
	// ProjectColored indicates whether project uses color (vs dim)
	ProjectColored bool
}

// AllContrastLevels is the registry of available contrast levels.
// Levels will cycle in order: medium → high → low → medium
var AllContrastLevels = []ContrastLevel{
	{
		Name:           "medium",
		DisplayName:    "Medium",
		AgentBold:      true,
		AgentColored:   true,
		ProjectColored: false,
	},
	{
		Name:           "high",
		DisplayName:    "High",
		AgentBold:      true,
		AgentColored:   true,
		ProjectColored: true,
	},
	{
		Name:           "low",
		DisplayName:    "Low",
		AgentBold:      false,
		AgentColored:   false,
		ProjectColored: false,
	},
}

// GetContrastLevel returns the contrast level with the given name, or the default if not found.
// Default contrast level is "medium".
func GetContrastLevel(name string) *ContrastLevel {
	for i := range AllContrastLevels {
		if AllContrastLevels[i].Name == name {
			return &AllContrastLevels[i]
		}
	}
	// Return default contrast level (first one)
	return &AllContrastLevels[0]
}

// NextContrastLevel returns the name of the next contrast level for cycling.
// If current contrast level is not found or is the last one, returns the first contrast level.
func NextContrastLevel(current string) string {
	for i, cl := range AllContrastLevels {
		if cl.Name == current {
			// Return next contrast level, wrapping around to first
			return AllContrastLevels[(i+1)%len(AllContrastLevels)].Name
		}
	}
	// If not found, return first contrast level
	return AllContrastLevels[0].Name
}
