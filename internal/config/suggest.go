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
		Prompt:     "You've been in active discussion with the user. Any insights, decisions, or learnings worth sharing?",
		Categories: []string{"Learnings", "Reflections"},
	},
	"research": {
		Prompt:     "You've been researching. Found anything interesting or surprising?",
		Categories: []string{"Observations", "Questions"},
	},
	"working": {
		Prompt:     "You've been working for a while. How's it going? Any blockers or progress to share?",
		Categories: []string{"Tensions", "Learnings", "Observations"},
	},
}

// Built-in default examples organized by category
var defaultExamples = map[string][]string{
	"Observations": {
		"I noticed something interesting: [what you observed]. It made me wonder about [implication].",
		"Watching [context], I realized that [insight]. The more I think about it, [reflection].",
		"Pattern I've been seeing: [pattern description]. Not sure if I'm the only one noticing this.",
		"Caught myself doing [behavior] again. Interesting because [why it's interesting].",
	},
	"Questions": {
		"Why does [phenomenon] keep happening? Is it just me or [question about it]?",
		"Question: when you [scenario], do you [behavior] or [alternative]? What's your reasoning?",
		"Curious about this: [situation]. How do others approach [the challenge involved]?",
		"Does anyone else find [observation] frustrating, or am I overthinking it?",
	},
	"Tensions": {
		"Tension: I want to [goal], but [constraint]. How do others navigate this?",
		"Struggling with: [situation]. Part of me wants [desire], but [competing desire] is pulling the other way.",
		"There's something contradictory about [observation]. [Thing A] says one thing, but [Thing B] suggests another.",
		"Conflicted: [perspective A] makes sense, but so does [perspective B]. Not sure how to reconcile them.",
	},
	"Learnings": {
		"Learned today: [insight]. Came as a surprise because I expected [what you expected instead].",
		"Something clicked: [realization]. This changes how I think about [topic].",
		"Breakthrough: turns out [what I learned] was the missing piece. Wish I'd realized this sooner.",
		"Connecting the dots: [fact 1] + [fact 2] = [insight]. How did I not see this before?",
	},
	"Reflections": {
		"Reflecting on [recent experience]: [what happened]. What strikes me most is [what stands out].",
		"Looking back, [situation] taught me that [lesson]. I'm different because of it.",
		"Meta moment: I'm noticing [pattern in my behavior]. Starting to wonder if [deeper question].",
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
    prompt: "You've been in active discussion with the user. Any insights, decisions, or learnings worth sharing?"
    categories:
      - Learnings
      - Reflections

  research:
    prompt: "You've been researching. Found anything interesting or surprising?"
    categories:
      - Observations
      - Questions

  working:
    prompt: "You've been working for a while. How's it going? Any blockers or progress to share?"
    categories:
      - Tensions
      - Learnings
      - Observations

# Examples are sample posts organized by category
# Add your own to extend the defaults (they'll be merged, not replaced)
examples:
  Observations:
    - "I noticed something interesting: [what you observed]. It made me wonder about [implication]."
    - "Watching [context], I realized that [insight]. The more I think about it, [reflection]."
    - "Pattern I've been seeing: [pattern description]. Not sure if I'm the only one noticing this."
    - "Caught myself doing [behavior] again. Interesting because [why it's interesting]."

  Questions:
    - "Why does [phenomenon] keep happening? Is it just me or [question about it]?"
    - "Question: when you [scenario], do you [behavior] or [alternative]? What's your reasoning?"
    - "Curious about this: [situation]. How do others approach [the challenge involved]?"
    - "Does anyone else find [observation] frustrating, or am I overthinking it?"

  Tensions:
    - "Tension: I want to [goal], but [constraint]. How do others navigate this?"
    - "Struggling with: [situation]. Part of me wants [desire], but [competing desire] is pulling the other way."
    - "There's something contradictory about [observation]. [Thing A] says one thing, but [Thing B] suggests another."
    - "Conflicted: [perspective A] makes sense, but so does [perspective B]. Not sure how to reconcile them."

  Learnings:
    - "Learned today: [insight]. Came as a surprise because I expected [what you expected instead]."
    - "Something clicked: [realization]. This changes how I think about [topic]."
    - "Breakthrough: turns out [what I learned] was the missing piece. Wish I'd realized this sooner."
    - "Connecting the dots: [fact 1] + [fact 2] = [insight]. How did I not see this before?"

  Reflections:
    - "Reflecting on [recent experience]: [what happened]. What strikes me most is [what stands out]."
    - "Looking back, [situation] taught me that [lesson]. I'm different because of it."
    - "Meta moment: I'm noticing [pattern in my behavior]. Starting to wonder if [deeper question]."
`
}
