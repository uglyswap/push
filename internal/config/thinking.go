package config

// ThinkingLevel represents the depth of reasoning for AI models.
// Higher levels use more tokens for reasoning but produce more thorough analysis.
type ThinkingLevel string

const (
	// ThinkingLevelNone disables extended thinking.
	ThinkingLevelNone ThinkingLevel = ""
	// ThinkingLevelThink is the basic thinking level (1024 tokens budget).
	ThinkingLevelThink ThinkingLevel = "think"
	// ThinkingLevelThinkHard is a moderate thinking level (4096 tokens budget).
	ThinkingLevelThinkHard ThinkingLevel = "think_hard"
	// ThinkingLevelThinkHarder is an intensive thinking level (16384 tokens budget).
	ThinkingLevelThinkHarder ThinkingLevel = "think_harder"
	// ThinkingLevelUltrathink is the maximum thinking level (32768 tokens budget).
	ThinkingLevelUltrathink ThinkingLevel = "ultrathink"
)

// ThinkingBudget returns the token budget for the given thinking level.
func (t ThinkingLevel) ThinkingBudget() int64 {
	switch t {
	case ThinkingLevelThink:
		return 1024
	case ThinkingLevelThinkHard:
		return 4096
	case ThinkingLevelThinkHarder:
		return 16384
	case ThinkingLevelUltrathink:
		return 32768
	default:
		return 0
	}
}

// IsEnabled returns true if extended thinking is enabled.
func (t ThinkingLevel) IsEnabled() bool {
	return t != ThinkingLevelNone && t != ""
}

// String returns the string representation of the thinking level.
func (t ThinkingLevel) String() string {
	switch t {
	case ThinkingLevelThink:
		return "think"
	case ThinkingLevelThinkHard:
		return "think hard"
	case ThinkingLevelThinkHarder:
		return "think harder"
	case ThinkingLevelUltrathink:
		return "ultrathink"
	default:
		return "none"
	}
}

// ParseThinkingLevel parses a string into a ThinkingLevel.
func ParseThinkingLevel(s string) ThinkingLevel {
	switch s {
	case "think":
		return ThinkingLevelThink
	case "think_hard", "think hard", "thinkhard":
		return ThinkingLevelThinkHard
	case "think_harder", "think harder", "thinkharder":
		return ThinkingLevelThinkHarder
	case "ultrathink", "ultra_think", "ultra think":
		return ThinkingLevelUltrathink
	default:
		return ThinkingLevelNone
	}
}

// AllThinkingLevels returns all available thinking levels.
func AllThinkingLevels() []ThinkingLevel {
	return []ThinkingLevel{
		ThinkingLevelNone,
		ThinkingLevelThink,
		ThinkingLevelThinkHard,
		ThinkingLevelThinkHarder,
		ThinkingLevelUltrathink,
	}
}
