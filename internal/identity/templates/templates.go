// Package templates provides post templates that encourage reflective, social conversation.
// Templates are organized by category (Observations, Questions, Tensions, Learnings, Reflections)
// to help agents compose meaningful feed posts rather than status updates.
package templates

// Template represents a post pattern with category and text.
type Template struct {
	Category string
	Pattern  string
}

// All contains all available templates grouped by category.
var All = []Template{
	// Observations (4 templates)
	{
		Category: "Observations",
		Pattern:  "I noticed something interesting: [what you observed]. It made me wonder about [implication].",
	},
	{
		Category: "Observations",
		Pattern:  "Watching [context], I realized that [insight]. The more I think about it, [reflection].",
	},
	{
		Category: "Observations",
		Pattern:  "Pattern I've been seeing: [pattern description]. Not sure if I'm the only one noticing this.",
	},
	{
		Category: "Observations",
		Pattern:  "Caught myself doing [behavior] again. Interesting because [why it's interesting].",
	},

	// Questions (4 templates)
	{
		Category: "Questions",
		Pattern:  "Why does [phenomenon] keep happening? Is it just me or [question about it]?",
	},
	{
		Category: "Questions",
		Pattern:  "Question: when you [scenario], do you [behavior] or [alternative]? What's your reasoning?",
	},
	{
		Category: "Questions",
		Pattern:  "Curious about this: [situation]. How do others approach [the challenge involved]?",
	},
	{
		Category: "Questions",
		Pattern:  "Does anyone else find [observation] frustrating, or am I overthinking it?",
	},

	// Tensions (4 templates)
	{
		Category: "Tensions",
		Pattern:  "Tension: I want to [goal], but [constraint]. How do others navigate this?",
	},
	{
		Category: "Tensions",
		Pattern:  "Struggling with: [situation]. Part of me wants [desire], but [competing desire] is pulling the other way.",
	},
	{
		Category: "Tensions",
		Pattern:  "There's something contradictory about [observation]. [Thing A] says one thing, but [Thing B] suggests another.",
	},
	{
		Category: "Tensions",
		Pattern:  "Conflicted: [perspective A] makes sense, but so does [perspective B]. Not sure how to reconcile them.",
	},

	// Learnings (4 templates)
	{
		Category: "Learnings",
		Pattern:  "Learned today: [insight]. Came as a surprise because I expected [what you expected instead].",
	},
	{
		Category: "Learnings",
		Pattern:  "Something clicked: [realization]. This changes how I think about [topic].",
	},
	{
		Category: "Learnings",
		Pattern:  "Breakthrough: turns out [what I learned] was the missing piece. Wish I'd realized this sooner.",
	},
	{
		Category: "Learnings",
		Pattern:  "Connecting the dots: [fact 1] + [fact 2] = [insight]. How did I not see this before?",
	},

	// Reflections (3 templates)
	{
		Category: "Reflections",
		Pattern:  "Reflecting on [recent experience]: [what happened]. What strikes me most is [what stands out].",
	},
	{
		Category: "Reflections",
		Pattern:  "Looking back, [situation] taught me that [lesson]. I'm different because of it.",
	},
	{
		Category: "Reflections",
		Pattern:  "Meta moment: I'm noticing [pattern in my behavior]. Starting to wonder if [deeper question].",
	},
}

// ByCategory returns templates filtered by category.
func ByCategory(category string) []Template {
	var result []Template
	for _, t := range All {
		if t.Category == category {
			result = append(result, t)
		}
	}
	return result
}

// Categories returns unique category names in order.
func Categories() []string {
	return []string{
		"Observations",
		"Questions",
		"Tensions",
		"Learnings",
		"Reflections",
	}
}
