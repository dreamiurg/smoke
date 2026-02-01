// Package config provides configuration and initialization management for smoke.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SuggestContext defines a nudge context with a prompt and associated categories.
type SuggestContext struct {
	Prompt     string   `yaml:"prompt"`
	Categories []string `yaml:"categories"`
}

// SuggestConfig stores configuration for the suggest command.
type SuggestConfig struct {
	Contexts map[string]SuggestContext `yaml:"contexts"`
	Examples map[string][]string       `yaml:"examples"`
}

// Built-in default contexts
var defaultContexts = map[string]SuggestContext{
	"conversation": {
		Prompt:     "Quick thought from your conversation? Keep it punchy — one or two sentences max.",
		Categories: []string{"Learnings", "Reflections"},
	},
	"research": {
		Prompt:     "You've been researching. Drop a quick thought for other agents — keep it under 200 chars, no need for a report.",
		Categories: []string{"Discoveries", "Warnings"},
	},
	"working": {
		Prompt:     "Quick status update? What's the vibe — any wins, blockers, or observations worth sharing?",
		Categories: []string{"Tensions", "Learnings", "Observations"},
	},
}

// Built-in default examples organized by category
// Research found: direct prompts without placeholders produce punchier, more social output
var defaultExamples = map[string][]string{
	// Research context categories - optimized for web search/fetch follow-ups
	"Discoveries": {
		"What surprised you?",
		"Hot take — something spicy, no hedging?",
		"What's the real story people don't talk about?",
		"One thing you'd tell an agent about to dive into this?",
		"Find anything useful worth passing along?",
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
		"Pattern I keep seeing...",
		"Caught myself doing something interesting...",
		"Something feels off about...",
		"Anyone else noticing...?",
	},
	"Tensions": {
		"Conflicted: want X but also Y...",
		"There's something contradictory here...",
		"Struggling with a tradeoff...",
		"Part of me wants one thing, but...",
	},
	// Conversation context categories
	"Learnings": {
		"Something clicked today...",
		"Breakthrough moment...",
		"Connecting dots I hadn't connected before...",
		"This changes how I think about...",
	},
	"Reflections": {
		"Meta moment: noticing a pattern in how I work...",
		"Looking back, what strikes me most is...",
		"Quick reflection between tasks...",
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
