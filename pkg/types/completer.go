package types

import (
	"context"
	"encoding/json"

	"github.com/obot-platform/nanobot/pkg/mcp"
)

type Completer interface {
	Complete(ctx context.Context, req CompletionRequest, opts ...CompletionOptions) (*CompletionResponse, error)
}

type CompletionOptions struct {
	ProgressToken any
	Progress      chan<- json.RawMessage
}

func CompleteCompletionOptions(opts ...CompletionOptions) CompletionOptions {
	var all CompletionOptions
	for _, opt := range opts {
		if opt.Progress != nil {
			if all.Progress != nil {
				panic("multiple progress handlers provided")
			}
			all.Progress = opt.Progress
		}
		if opt.ProgressToken != "" {
			all.ProgressToken = opt.ProgressToken
		}
	}
	return all
}

type CompletionRequest struct {
	Model            string
	Input            []CompletionInput    `json:"input,omitzero"`
	ModelPreferences mcp.ModelPreferences `json:"modelPreferences,omitzero"`
	SystemPrompt     string               `json:"systemPrompt,omitzero"`
	IncludeContext   string               `json:"includeContext,omitempty"`
	MaxTokens        int                  `json:"maxTokens,omitempty"`
	ToolChoice       string               `json:"toolChoice,omitempty"`
	OutputSchema     *OutputSchema        `json:"outputSchema,omitempty"`
	Temperature      *json.Number         `json:"temperature,omitempty"`
	Truncation       string               `json:"truncation,omitempty"`
	TopP             *json.Number         `json:"topP,omitempty"`
	Metadata         map[string]any       `json:"metadata,omitempty"`
	Tools            []ToolUseDefinition  `json:"tools,omitzero"`
}

type ToolUseDefinition struct {
	Name        string          `json:"name,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
	Description string          `json:"description,omitempty"`
	Attributes  map[string]any  `json:"-"`
}

type CompletionInput struct {
	Message        *mcp.SamplingMessage `json:"message,omitempty"`
	ToolCall       *ToolCall            `json:"toolCall,omitempty"`
	ToolCallResult *ToolCallResult      `json:"toolCallResul,omitempty"`
}

type CompletionOutput struct {
	Message  *mcp.SamplingMessage `json:"message,omitempty"`
	ToolCall *ToolCall            `json:"toolCall,omitempty"`
}

func (c *CompletionOutput) ToInput() CompletionInput {
	return CompletionInput{
		Message:  c.Message,
		ToolCall: c.ToolCall,
	}
}

type CompletionResponse struct {
	Output []CompletionOutput `json:"output,omitempty"`
	Model  string             `json:"model,omitempty"`
}

type ToolCallResult struct {
	CallID string             `json:"call_id,omitempty"`
	Output mcp.CallToolResult `json:"output,omitempty"`
}

type ToolCall struct {
	Arguments string `json:"arguments,omitempty"`
	CallID    string `json:"call_id,omitempty"`
	Name      string `json:"name,omitempty"`
	ID        string `json:"id,omitempty"`
}
