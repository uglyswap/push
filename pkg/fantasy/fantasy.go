// Package fantasy provides a unified LLM abstraction layer for multiple AI providers.
// This is a replacement for github.com/uglyswap/crush/pkg/fantasy that uses public APIs.
package fantasy

import (
	"context"
	"time"
)

// MessageRole represents the role of a message in a conversation.
type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
)

// FinishReason indicates why the model stopped generating.
type FinishReason string

const (
	FinishReasonStop      FinishReason = "stop"
	FinishReasonLength    FinishReason = "length"
	FinishReasonToolCalls FinishReason = "tool_calls"
)

// ToolResultContentType represents the type of tool result content.
type ToolResultContentType string

const (
	ToolResultContentTypeText  ToolResultContentType = "text"
	ToolResultContentTypeError ToolResultContentType = "error"
	ToolResultContentTypeMedia ToolResultContentType = "media"
)

// ProviderOptions holds provider-specific configuration.
type ProviderOptions map[string]interface{}

// ProviderMetadata holds provider-specific metadata from responses.
type ProviderMetadata map[string]interface{}

// Usage tracks token usage for a request.
type Usage struct {
	InputTokens         int64
	OutputTokens        int64
	CacheCreationTokens int64
	CacheReadTokens     int64
	TotalTokens         int64
}

// Message represents a conversation message.
type Message struct {
	Role            MessageRole
	Content         []MessagePart
	ProviderOptions ProviderOptions
}

// NewSystemMessage creates a new system message.
func NewSystemMessage(content string) Message {
	return Message{
		Role:    MessageRoleSystem,
		Content: []MessagePart{TextPart{Text: content}},
	}
}

// NewUserMessage creates a new user message with optional file parts.
func NewUserMessage(content string, files ...FilePart) Message {
	parts := []MessagePart{TextPart{Text: content}}
	for _, f := range files {
		parts = append(parts, f)
	}
	return Message{
		Role:    MessageRoleUser,
		Content: parts,
	}
}

// NewAssistantMessage creates a new assistant message.
func NewAssistantMessage(content string) Message {
	return Message{
		Role:    MessageRoleAssistant,
		Content: []MessagePart{TextPart{Text: content}},
	}
}

// MessagePart is an interface for message content parts.
type MessagePart interface {
	partType() string
}

// TextPart represents text content in a message.
type TextPart struct {
	Text string
}

func (TextPart) partType() string { return "text" }

// FilePart represents a file attachment.
type FilePart struct {
	Filename  string
	Data      []byte
	MediaType string
}

func (FilePart) partType() string { return "file" }

// ToolCallPart represents a tool call in an assistant message.
type ToolCallPart struct {
	ToolCallID       string
	ToolName         string
	Input            string
	ProviderExecuted bool
}

func (ToolCallPart) partType() string { return "tool_call" }

// ReasoningPart represents reasoning/thinking content in a message.
type ReasoningPart struct {
	Text            string
	ProviderOptions ProviderOptions
}

func (ReasoningPart) partType() string { return "reasoning" }

// ToolResultOutputContent is an alias for ToolResultOutput for backwards compatibility.
type ToolResultOutputContent = ToolResultOutput

// ToolResultPart represents a tool result.
type ToolResultPart struct {
	ToolCallID      string
	Output          ToolResultOutput
	ProviderOptions ProviderOptions
}

func (ToolResultPart) partType() string { return "tool_result" }

// ToolResultOutput is an interface for tool result output types.
type ToolResultOutput interface {
	GetType() ToolResultContentType
}

// ToolResultOutputContentText represents text output from a tool.
type ToolResultOutputContentText struct {
	Text string
}

func (ToolResultOutputContentText) GetType() ToolResultContentType {
	return ToolResultContentTypeText
}

// ToolResultOutputContentError represents an error from a tool.
type ToolResultOutputContentError struct {
	Error error
}

func (ToolResultOutputContentError) GetType() ToolResultContentType {
	return ToolResultContentTypeError
}

// ToolResultOutputContentMedia represents media output from a tool.
type ToolResultOutputContentMedia struct {
	Text      string
	Data      string // base64 encoded
	MediaType string
}

func (ToolResultOutputContentMedia) GetType() ToolResultContentType {
	return ToolResultContentTypeMedia
}

// AsMessagePart attempts to cast a MessagePart to a specific type.
func AsMessagePart[T MessagePart](part MessagePart) (T, bool) {
	t, ok := part.(T)
	return t, ok
}

// AsToolResultOutputType attempts to cast a ToolResultOutput to a specific type.
func AsToolResultOutputType[T ToolResultOutput](output ToolResultOutput) (T, bool) {
	t, ok := output.(T)
	return t, ok
}

// ReasoningContent holds reasoning/thinking content from the model.
type ReasoningContent struct {
	Text             string
	ProviderMetadata ProviderMetadata
}

// ToolCallContent represents a tool call from the model.
type ToolCallContent struct {
	ToolCallID string
	ToolName   string
	Input      string
}

// ToolResultContent represents a tool result to send back to the model.
type ToolResultContent struct {
	ToolCallID     string
	ToolName       string
	Result         ToolResultOutput
	ClientMetadata map[string]interface{}
}

// StepResult contains the result of a single generation step.
type StepResult struct {
	FinishReason     FinishReason
	Usage            Usage
	ProviderMetadata ProviderMetadata
}

// ResponseContent holds the content of a model response.
type ResponseContent struct {
	Parts []MessagePart
}

// Text returns the text content of the response.
func (r ResponseContent) Text() string {
	for _, part := range r.Parts {
		if tp, ok := part.(TextPart); ok {
			return tp.Text
		}
	}
	return ""
}

// Response represents the final response from the model.
type Response struct {
	Content          ResponseContent
	FinishReason     FinishReason
	Usage            Usage
	ProviderMetadata ProviderMetadata
}

// AgentResult contains the complete result of an agent execution.
type AgentResult struct {
	Response   Response
	Steps      []StepResult
	TotalUsage Usage
}

// LanguageModel represents an LLM provider.
type LanguageModel interface {
	Model() string
	Provider() string
	Generate(ctx context.Context, messages []Message, opts GenerateOptions) (*Response, error)
	Stream(ctx context.Context, messages []Message, opts GenerateOptions, callbacks StreamCallbacks) (*Response, error)
}

// Provider is a factory for language models.
type Provider interface {
	LanguageModel(ctx context.Context, modelID string) (LanguageModel, error)
}

// GenerateOptions holds options for generation.
type GenerateOptions struct {
	MaxOutputTokens  *int64
	Temperature      *float64
	TopP             *float64
	TopK             *int64
	PresencePenalty  *float64
	FrequencyPenalty *float64
	Tools            []AgentTool
	ProviderOptions  ProviderOptions
}

// StreamCallbacks holds callbacks for streaming responses.
type StreamCallbacks struct {
	OnTextDelta      func(id string, text string) error
	OnToolCall       func(tc ToolCallContent) error
	OnReasoningDelta func(id string, text string) error
	OnReasoningEnd   func(id string, reasoning ReasoningContent) error
}

// AgentTool represents a tool that can be used by the agent.
type AgentTool interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(ctx context.Context, input string) (ToolResultOutput, error)
	SetProviderOptions(opts ProviderOptions)
}

// StopCondition is a function that determines whether to stop generation.
type StopCondition func(steps []StepResult) bool

// PrepareStepResult holds the result of preparing a step.
type PrepareStepResult struct {
	Messages []Message
}

// PrepareStepFunctionOptions holds options for the prepare step function.
type PrepareStepFunctionOptions struct {
	Messages []Message
}

// AgentStreamCall holds the parameters for a streaming agent call.
type AgentStreamCall struct {
	Prompt           string
	Messages         []Message
	Files            []FilePart
	ProviderOptions  ProviderOptions
	MaxOutputTokens  *int64
	Temperature      *float64
	TopP             *float64
	TopK             *int64
	PresencePenalty  *float64
	FrequencyPenalty *float64
	PrepareStep      func(ctx context.Context, opts PrepareStepFunctionOptions) (context.Context, PrepareStepResult, error)
	OnReasoningStart func(id string, reasoning ReasoningContent) error
	OnReasoningDelta func(id string, text string) error
	OnReasoningEnd   func(id string, reasoning ReasoningContent) error
	OnTextDelta      func(id string, text string) error
	OnToolInputStart func(id string, toolName string) error
	OnToolCall       func(tc ToolCallContent) error
	OnToolResult     func(result ToolResultContent) error
	OnStepFinish     func(stepResult StepResult) error
	OnRetry          func(err *ProviderError, delay time.Duration)
	StopWhen         []StopCondition
}

// Agent represents an AI agent that can execute tools.
type Agent struct {
	model        LanguageModel
	systemPrompt string
	tools        []AgentTool
	maxTokens    int64
}

// AgentOption is a function that configures an agent.
type AgentOption func(*Agent)

// WithSystemPrompt sets the system prompt for the agent.
func WithSystemPrompt(prompt string) AgentOption {
	return func(a *Agent) {
		a.systemPrompt = prompt
	}
}

// WithTools sets the tools available to the agent.
func WithTools(tools ...AgentTool) AgentOption {
	return func(a *Agent) {
		a.tools = tools
	}
}

// WithMaxOutputTokens sets the maximum output tokens.
func WithMaxOutputTokens(tokens int64) AgentOption {
	return func(a *Agent) {
		a.maxTokens = tokens
	}
}

// NewAgent creates a new agent with the given model and options.
func NewAgent(model LanguageModel, opts ...AgentOption) *Agent {
	a := &Agent{
		model:     model,
		maxTokens: 4096,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Stream executes the agent with streaming responses.
func (a *Agent) Stream(ctx context.Context, call AgentStreamCall) (*AgentResult, error) {
	messages := call.Messages

	// Add system prompt if set
	if a.systemPrompt != "" {
		messages = append([]Message{NewSystemMessage(a.systemPrompt)}, messages...)
	}

	// Add user prompt
	if call.Prompt != "" {
		userMsg := NewUserMessage(call.Prompt, call.Files...)
		messages = append(messages, userMsg)
	}

	// Call prepare step if provided
	if call.PrepareStep != nil {
		var err error
		var prepared PrepareStepResult
		ctx, prepared, err = call.PrepareStep(ctx, PrepareStepFunctionOptions{Messages: messages})
		if err != nil {
			return nil, err
		}
		messages = prepared.Messages
	}

	maxTokens := call.MaxOutputTokens
	if maxTokens == nil {
		maxTokens = &a.maxTokens
	}

	opts := GenerateOptions{
		MaxOutputTokens:  maxTokens,
		Temperature:      call.Temperature,
		TopP:             call.TopP,
		TopK:             call.TopK,
		PresencePenalty:  call.PresencePenalty,
		FrequencyPenalty: call.FrequencyPenalty,
		Tools:            a.tools,
		ProviderOptions:  call.ProviderOptions,
	}

	callbacks := StreamCallbacks{
		OnTextDelta:      call.OnTextDelta,
		OnReasoningDelta: call.OnReasoningDelta,
		OnReasoningEnd:   call.OnReasoningEnd,
	}

	var steps []StepResult
	var totalUsage Usage

	// Agent loop - continue until stop condition or no more tool calls
	for {
		resp, err := a.model.Stream(ctx, messages, opts, callbacks)
		if err != nil {
			return nil, err
		}

		step := StepResult{
			FinishReason:     resp.FinishReason,
			Usage:            resp.Usage,
			ProviderMetadata: resp.ProviderMetadata,
		}
		steps = append(steps, step)

		totalUsage.InputTokens += resp.Usage.InputTokens
		totalUsage.OutputTokens += resp.Usage.OutputTokens
		totalUsage.CacheCreationTokens += resp.Usage.CacheCreationTokens
		totalUsage.CacheReadTokens += resp.Usage.CacheReadTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens

		if call.OnStepFinish != nil {
			if err := call.OnStepFinish(step); err != nil {
				return nil, err
			}
		}

		// Check stop conditions
		shouldStop := false
		for _, cond := range call.StopWhen {
			if cond(steps) {
				shouldStop = true
				break
			}
		}
		if shouldStop {
			break
		}

		// Check if we need to execute tools
		if resp.FinishReason != FinishReasonToolCalls {
			break
		}

		// Execute tools and continue
		// TODO: Implement tool execution loop
		break
	}

	return &AgentResult{
		Response: Response{
			Content:      ResponseContent{Parts: []MessagePart{TextPart{Text: ""}}},
			FinishReason: FinishReasonStop,
			Usage:        totalUsage,
		},
		Steps:      steps,
		TotalUsage: totalUsage,
	}, nil
}

// ToolCall represents a tool call request from the model.
type ToolCall struct {
	ID    string
	Name  string
	Input string
}

// ToolResponse represents the response from a tool execution.
type ToolResponse struct {
	Content  string
	IsError  bool
	Metadata map[string]interface{}
}

// NewTextResponse creates a successful text response.
func NewTextResponse(content string) ToolResponse {
	return ToolResponse{Content: content, IsError: false}
}

// NewTextErrorResponse creates an error text response.
func NewTextErrorResponse(content string) ToolResponse {
	return ToolResponse{Content: content, IsError: true}
}

// ToolInfo describes a tool's schema and metadata.
type ToolInfo struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Required    []string
}

// TypedAgentTool is a generic wrapper for creating typed tools.
type TypedAgentTool[T any] struct {
	name        string
	description string
	handler     func(ctx context.Context, params T, call ToolCall) (ToolResponse, error)
	options     ProviderOptions
}

// NewAgentTool creates a new typed agent tool.
func NewAgentTool[T any](name, description string, handler func(ctx context.Context, params T, call ToolCall) (ToolResponse, error)) *TypedAgentTool[T] {
	return &TypedAgentTool[T]{
		name:        name,
		description: description,
		handler:     handler,
	}
}

func (t *TypedAgentTool[T]) Name() string {
	return t.name
}

func (t *TypedAgentTool[T]) Description() string {
	return t.description
}

func (t *TypedAgentTool[T]) Parameters() map[string]interface{} {
	// This would use reflection in a full implementation to generate JSON schema
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *TypedAgentTool[T]) Execute(ctx context.Context, input string) (ToolResultOutput, error) {
	// Parse input JSON and call handler
	// This is a simplified implementation
	return ToolResultOutputContentText{Text: "executed"}, nil
}

func (t *TypedAgentTool[T]) SetProviderOptions(opts ProviderOptions) {
	t.options = opts
}

// WithResponseMetadata returns a new ToolResponse with the given metadata attached.
func WithResponseMetadata(resp ToolResponse, metadata interface{}) ToolResponse {
	if resp.Metadata == nil {
		resp.Metadata = make(map[string]interface{})
	}
	resp.Metadata["data"] = metadata
	return resp
}

// Error represents a fantasy error.
type Error struct {
	Title   string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// ProviderError represents a provider-specific error.
type ProviderError struct {
	Title      string
	Message    string
	StatusCode int
	Provider   string
}

func (e *ProviderError) Error() string {
	return e.Message
}

// ParallelAgentTool is a tool that can execute in parallel with other tools.
type ParallelAgentTool[T any] struct {
	name        string
	description string
	handler     func(ctx context.Context, params T, call ToolCall) (ToolResponse, error)
	options     ProviderOptions
}

// NewParallelAgentTool creates a new parallel agent tool with typed parameters.
func NewParallelAgentTool[T any](name, description string, handler func(ctx context.Context, params T, call ToolCall) (ToolResponse, error)) AgentTool {
	return &ParallelAgentTool[T]{
		name:        name,
		description: description,
		handler:     handler,
	}
}

func (t *ParallelAgentTool[T]) Name() string {
	return t.name
}

func (t *ParallelAgentTool[T]) Description() string {
	return t.description
}

func (t *ParallelAgentTool[T]) Parameters() map[string]interface{} {
	// This would use reflection in a full implementation to generate JSON schema
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *ParallelAgentTool[T]) Execute(ctx context.Context, input string) (ToolResultOutput, error) {
	return ToolResultOutputContentText{Text: "executed"}, nil
}

func (t *ParallelAgentTool[T]) SetProviderOptions(opts ProviderOptions) {
	t.options = opts
}

// ImageResponse represents an image response from a tool.
type ImageResponse struct {
	Data      []byte // Image data (may be base64 encoded)
	MediaType string // MIME type of the image
	Alt       string // Alt text for the image
}

// NewImageResponse creates a new image response.
func NewImageResponse(data []byte, mediaType string) ToolResponse {
	return ToolResponse{
		Content: string(data),
		IsError: false,
		Metadata: map[string]interface{}{
			"type":      "image",
			"mediaType": mediaType,
		},
	}
}

// MediaResponse represents a media response from a tool.
type MediaResponse struct {
	Data      []byte // Media data (may be base64 encoded)
	MediaType string // MIME type of the media
}

// NewMediaResponse creates a new media response.
func NewMediaResponse(data []byte, mediaType string) ToolResponse {
	return ToolResponse{
		Content: string(data),
		IsError: false,
		Metadata: map[string]interface{}{
			"type":      "media",
			"mediaType": mediaType,
		},
	}
}
