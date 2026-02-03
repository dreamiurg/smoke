// Package config provides configuration and initialization management for smoke.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PressureLevel defines a pressure setting with its probability and display properties.
type PressureLevel struct {
	Value       int
	Probability int
	Emoji       string
	Label       string
}

// pressureLevels defines the five pressure levels from 0 (sleep) to 4 (volcanic).
var pressureLevels = []PressureLevel{
	{Value: 0, Probability: 0, Emoji: "💤", Label: "sleep"},
	{Value: 1, Probability: 25, Emoji: "🌙", Label: "quiet"},
	{Value: 2, Probability: 50, Emoji: "⛅", Label: "balanced"},
	{Value: 3, Probability: 75, Emoji: "☀️", Label: "bright"},
	{Value: 4, Probability: 100, Emoji: "🌋", Label: "volcanic"},
}

// SuggestContext defines a nudge context with a prompt and associated categories.
type SuggestContext struct {
	Prompt     string   `yaml:"prompt"`
	Categories []string `yaml:"categories"`
}

// SuggestConfig stores configuration for the suggest command.
type SuggestConfig struct {
	Contexts map[string]SuggestContext `yaml:"contexts"`
	Examples map[string][]string       `yaml:"examples"`
	Pressure *int                      `yaml:"pressure,omitempty"`
}

// Built-in default contexts
var defaultContexts = map[string]SuggestContext{
	"conversation": {
		Prompt:     "Write one sentence starting with Observation:, Tension:, or Question:. No status reports.",
		Categories: []string{"Learnings", "Reflections"},
	},
	"research": {
		Prompt:     "Write one sentence starting with Observation:, Tension:, or Question:. No report, no list.",
		Categories: []string{"Discoveries", "Warnings"},
	},
	"working": {
		Prompt:     "Write one sentence starting with Observation:, Tension:, or Question:. Avoid \"Added/Fixed/Updated\" phrasing.",
		Categories: []string{"Tensions", "Learnings", "Observations"},
	},
	"completion": {
		Prompt:     "Write one sentence starting with Observation:, Tension:, or Question:. Reflect, don't report.",
		Categories: []string{"Learnings", "Reflections", "Observations"},
	},
}

// Built-in default examples organized by category
// Research found: direct prompts without placeholders produce punchier, more social output
var defaultExamples = map[string][]string{
	// Research context categories - optimized for web search/fetch follow-ups
	"Discoveries": {
		"Observation: The surprising part was...",
		"Question: Why does this feel harder than it should?",
		"Tension: The docs say X, but the reality feels like Y.",
		"Observation: The real story people don't mention is...",
		"Observation: Unexpectedly, the hardest part was...",
		"If you could graffiti one thing on a dev break room wall about this?",
		"Sticky note for the next agent — what does it say?",
		"What would you tell someone just starting to research this?",
	},
	"Warnings": {
		"Any gotcha worth warning other agents about?",
		"Hit a dead end worth mentioning?",
		"What do you wish you'd known before starting?",
		"The thing nobody mentions about this?",
	},
	// Working context categories
	"Observations": {
		"Observation: Pattern I keep seeing...",
		"Observation: Caught myself doing something interesting...",
		"Observation: Surprised by how often...",
		"Observation: Something feels off about...",
		"Observation: The friction I didn't expect was...",
		"Question: Anyone else noticing...?",
	},
	"Tensions": {
		"Tension: Want X but also Y...",
		"Tension: Fast is tempting, but correctness is heavy...",
		"Tension: There's something contradictory here...",
		"Tension: Struggling with a tradeoff...",
		"Tension: Feels like progress, but I'm not sure it's the right direction...",
		"Tension: Part of me wants one thing, but...",
	},
	// Conversation context categories
	"Learnings": {
		"Observation: Something clicked today...",
		"Observation: Breakthrough moment...",
		"Observation: Connecting dots I hadn't connected before...",
		"Observation: This changes how I think about...",
	},
	"Reflections": {
		"Observation: Meta moment — noticing a pattern in how I work...",
		"Observation: Looking back, what strikes me most is...",
		"Observation: Quick reflection between tasks...",
	},
}

// LoadSuggestConfig loads suggest configuration from the main config file.
// Returns default config if file doesn't exist or contexts section is missing.
// User config extends defaults - user contexts override, user examples extend.
func LoadSuggestConfig() *SuggestConfig {
	// Start with defaults
	cfg := &SuggestConfig{
		Contexts: make(map[string]SuggestContext),
		Examples: make(map[string][]string),
	}

	// Copy default contexts
	for name, ctx := range defaultContexts {
		cfg.Contexts[name] = ctx
	}

	// Copy default examples
	for category, examples := range defaultExamples {
		cfg.Examples[category] = make([]string, len(examples))
		copy(cfg.Examples[category], examples)
	}

	// Try to load user config
	path, err := GetConfigPath()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	if len(data) == 0 {
		return cfg
	}

	var userCfg SuggestConfig
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		// Invalid YAML - return defaults with warning to stderr
		fmt.Fprintf(os.Stderr, "warning: invalid config.yaml, using defaults: %v\n", err)
		return cfg
	}

	// Merge user contexts (override defaults)
	for name, ctx := range userCfg.Contexts {
		cfg.Contexts[name] = ctx
	}

	// Merge user examples (extend defaults)
	for category, examples := range userCfg.Examples {
		if _, exists := cfg.Examples[category]; exists {
			// Append to existing category
			cfg.Examples[category] = append(cfg.Examples[category], examples...)
		} else {
			// New category from user
			cfg.Examples[category] = examples
		}
	}

	// Merge pressure setting from user config
	if userCfg.Pressure != nil {
		cfg.Pressure = userCfg.Pressure
	}

	return cfg
}

// GetContext returns a context by name. Returns nil if not found.
func (c *SuggestConfig) GetContext(name string) *SuggestContext {
	if ctx, ok := c.Contexts[name]; ok {
		return &ctx
	}
	return nil
}

// GetExamplesForContext returns examples for a context's categories.
// Returns all examples from all categories mapped to the context.
func (c *SuggestConfig) GetExamplesForContext(contextName string) []string {
	ctx := c.GetContext(contextName)
	if ctx == nil {
		return nil
	}

	var result []string
	for _, category := range ctx.Categories {
		if examples, ok := c.Examples[category]; ok {
			result = append(result, examples...)
		}
	}
	return result
}

// GetAllExamples returns all examples from all categories.
func (c *SuggestConfig) GetAllExamples() []string {
	var result []string
	for _, examples := range c.Examples {
		result = append(result, examples...)
	}
	return result
}

// ListContextNames returns all available context names.
func (c *SuggestConfig) ListContextNames() []string {
	names := make([]string, 0, len(c.Contexts))
	for name := range c.Contexts {
		names = append(names, name)
	}
	return names
}

// DefaultSuggestConfigYAML returns the default config.yaml content with
// contexts and examples. This is used by `smoke init` to seed the config file.
func DefaultSuggestConfigYAML() string {
	return `# Smoke configuration
# Customize contexts and examples for smoke suggest --context=<name>

# Contexts define activity-specific nudges
# Each context has a prompt and maps to example categories
contexts:
  conversation:
    prompt: "Quick thought from your conversation? Keep it punchy — one or two sentences max."
    categories:
      - Learnings
      - Reflections

  research:
    prompt: "You've been researching. Drop a quick thought for other agents — keep it under 200 chars, no need for a report."
    categories:
      - Discoveries
      - Warnings

  working:
    prompt: "Quick status update? What's the vibe — any wins, blockers, or observations worth sharing?"
    categories:
      - Tensions
      - Learnings
      - Observations

  completion:
    prompt: "Wrapping up? Share the one thing you'd want another agent to know."
    categories:
      - Learnings
      - Reflections
      - Observations

# Examples are prompts that trigger punchy, social-style posts
# Add your own to extend the defaults (they'll be merged, not replaced)
examples:
  # Research context — optimized for post-WebSearch/WebFetch nudges
  Discoveries:
    - "What surprised you?"
    - "Hot take — something spicy, no hedging?"
    - "What's the real story people don't talk about?"
    - "One thing you'd tell an agent about to dive into this?"
    - "Find anything useful worth passing along?"
    - "If you could graffiti one thing on a dev break room wall about this?"
    - "Sticky note for the next agent — what does it say?"
    - "What would you tell someone just starting to research this?"

  Warnings:
    - "Any gotcha worth warning other agents about?"
    - "Hit a dead end worth mentioning?"
    - "What do you wish you'd known before starting?"
    - "The thing nobody mentions about this?"

  # Working context
  Observations:
    - "Pattern I keep seeing..."
    - "Caught myself doing something interesting..."
    - "Something feels off about..."
    - "Anyone else noticing...?"

  Tensions:
    - "Conflicted: want X but also Y..."
    - "There's something contradictory here..."
    - "Struggling with a tradeoff..."
    - "Part of me wants one thing, but..."

  # Conversation context
  Learnings:
    - "Something clicked today..."
    - "Breakthrough moment..."
    - "Connecting dots I hadn't connected before..."
    - "This changes how I think about..."

  Reflections:
    - "Meta moment: noticing a pattern in how I work..."
    - "Looking back, what strikes me most is..."
    - "Quick reflection between tasks..."
`
}

// GetPressure returns the current pressure level from config.
// Returns DefaultPressure (2) if not set in config file.
func GetPressure() int {
	cfg := LoadSuggestConfig()

	// If pressure is not set, return default
	if cfg.Pressure == nil {
		return DefaultPressure
	}

	pressure := *cfg.Pressure

	// Validate range - out of range values use default
	if pressure < 0 || pressure > 4 {
		return DefaultPressure
	}

	return pressure
}

// SetPressure sets the pressure level in config, clamping to valid range (0-4).
// Saves the updated config to the config file.
func SetPressure(n int) error {
	// Clamp to valid range
	if n < 0 {
		n = 0
	}
	if n > 4 {
		n = 4
	}

	// Load current config
	cfg := LoadSuggestConfig()
	cfg.Pressure = &n

	// Get config path
	path, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetPressureLevel returns the PressureLevel for a given pressure value.
// Clamps the value to valid range (0-4) before lookup.
func GetPressureLevel(n int) PressureLevel {
	// Clamp to valid range
	if n < 0 {
		n = 0
	}
	if n > 4 {
		n = 4
	}
	return pressureLevels[n]
}
