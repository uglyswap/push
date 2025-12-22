package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// AskUserQuestionTool allows the agent to ask the user questions.
type AskUserQuestionTool struct {
	// QuestionHandler is called when the agent asks a question.
	// It should return the user's response.
	QuestionHandler func(questions []Question) (map[string]string, error)
}

// Question represents a single question to ask the user.
type Question struct {
	Question    string   `json:"question"`
	Header      string   `json:"header"`
	Options     []Option `json:"options"`
	MultiSelect bool     `json:"multiSelect"`
}

// Option represents a selectable option for a question.
type Option struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// NewAskUserQuestionTool creates a new AskUserQuestion tool.
func NewAskUserQuestionTool(handler func(questions []Question) (map[string]string, error)) *AskUserQuestionTool {
	return &AskUserQuestionTool{
		QuestionHandler: handler,
	}
}

// Name returns the tool name.
func (t *AskUserQuestionTool) Name() string {
	return "AskUserQuestion"
}

// Description returns the tool description.
func (t *AskUserQuestionTool) Description() string {
	return `Use this tool when you need to ask the user questions during execution. This allows you to:
1. Gather user preferences or requirements
2. Clarify ambiguous instructions
3. Get decisions on implementation choices as you work
4. Offer choices to the user about what direction to take.

Usage notes:
- Users will always be able to select "Other" to provide custom text input
- Use multiSelect: true to allow multiple answers to be selected for a question
- If you recommend a specific option, make that the first option in the list and add "(Recommended)" at the end of the label`
}

// AskUserQuestionParams represents the parameters for AskUserQuestion.
type AskUserQuestionParams struct {
	Questions []Question        `json:"questions"`
	Answers   map[string]string `json:"answers,omitempty"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *AskUserQuestionTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"questions": map[string]interface{}{
				"type":        "array",
				"description": "Questions to ask the user (1-4 questions)",
				"minItems":    1,
				"maxItems":    4,
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"question": map[string]interface{}{
							"type":        "string",
							"description": "The complete question to ask the user",
						},
						"header": map[string]interface{}{
							"type":        "string",
							"description": "Very short label displayed as a chip/tag (max 12 chars)",
						},
						"options": map[string]interface{}{
							"type":        "array",
							"description": "Available choices (2-4 options)",
							"minItems":    2,
							"maxItems":    4,
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"label": map[string]interface{}{
										"type":        "string",
										"description": "Display text for this option (1-5 words)",
									},
									"description": map[string]interface{}{
										"type":        "string",
										"description": "Explanation of what this option means",
									},
								},
								"required": []string{"label", "description"},
							},
						},
						"multiSelect": map[string]interface{}{
							"type":        "boolean",
							"description": "Allow multiple selections",
						},
					},
					"required": []string{"question", "header", "options", "multiSelect"},
				},
			},
			"answers": map[string]interface{}{
				"type":        "object",
				"description": "User answers collected by the permission component",
			},
		},
		"required": []string{"questions"},
	}
}

// Execute runs the AskUserQuestion tool.
func (t *AskUserQuestionTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p AskUserQuestionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if len(p.Questions) == 0 {
		return "", fmt.Errorf("at least one question is required")
	}

	if len(p.Questions) > 4 {
		return "", fmt.Errorf("maximum 4 questions allowed")
	}

	// Validate questions
	for i, q := range p.Questions {
		if q.Question == "" {
			return "", fmt.Errorf("question %d: question text is required", i+1)
		}
		if q.Header == "" {
			return "", fmt.Errorf("question %d: header is required", i+1)
		}
		if len(q.Header) > 12 {
			return "", fmt.Errorf("question %d: header must be 12 characters or less", i+1)
		}
		if len(q.Options) < 2 || len(q.Options) > 4 {
			return "", fmt.Errorf("question %d: must have 2-4 options", i+1)
		}
	}

	// If answers are already provided (from UI), return them
	if len(p.Answers) > 0 {
		var result strings.Builder
		result.WriteString("User responses:\n\n")
		for i, q := range p.Questions {
			answer, ok := p.Answers[fmt.Sprintf("q%d", i)]
			if !ok {
				answer = p.Answers[q.Header]
			}
			if answer == "" {
				answer = "(no answer)"
			}
			result.WriteString(fmt.Sprintf("**%s**: %s\n", q.Header, answer))
		}
		return result.String(), nil
	}

	// Call the question handler if set
	if t.QuestionHandler != nil {
		answers, err := t.QuestionHandler(p.Questions)
		if err != nil {
			return "", fmt.Errorf("failed to get user response: %w", err)
		}

		var result strings.Builder
		result.WriteString("User responses:\n\n")
		for i, q := range p.Questions {
			answer, ok := answers[fmt.Sprintf("q%d", i)]
			if !ok {
				answer = answers[q.Header]
			}
			if answer == "" {
				answer = "(no answer)"
			}
			result.WriteString(fmt.Sprintf("**%s**: %s\n", q.Header, answer))
		}
		return result.String(), nil
	}

	// Format questions for display (when no handler is set)
	var result strings.Builder
	result.WriteString("Questions for user:\n\n")

	for i, q := range p.Questions {
		result.WriteString(fmt.Sprintf("### Question %d: %s\n", i+1, q.Header))
		result.WriteString(fmt.Sprintf("%s\n\n", q.Question))

		if q.MultiSelect {
			result.WriteString("(Select all that apply)\n")
		}

		for j, opt := range q.Options {
			result.WriteString(fmt.Sprintf("%d. **%s** - %s\n", j+1, opt.Label, opt.Description))
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *AskUserQuestionTool) RequiresApproval() bool {
	return true // Questions need user interaction
}
