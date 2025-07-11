package types

import (
	"context"
	"encoding/json"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

type Completer interface {
	Complete(ctx context.Context, req CompletionRequest, opts ...CompletionOptions) (*CompletionResponse, error)
}

type CompletionOptions struct {
	ProgressToken any
	Chat          *bool
}

func (c CompletionOptions) Merge(other CompletionOptions) (result CompletionOptions) {
	result.ProgressToken = complete.Last(c.ProgressToken, other.ProgressToken)
	result.Chat = complete.Last(c.Chat, other.Chat)
	return
}

type CompletionRequest struct {
	Model             string               `json:"model,omitempty"`
	ThreadName        string               `json:"threadName,omitempty"`
	NewThread         bool                 `json:"newThread,omitempty"`
	Input             []CompletionItem     `json:"input,omitzero"`
	ModelPreferences  mcp.ModelPreferences `json:"modelPreferences,omitzero"`
	SystemPrompt      string               `json:"systemPrompt,omitzero"`
	IncludeContext    string               `json:"includeContext,omitempty"`
	MaxTokens         int                  `json:"maxTokens,omitempty"`
	ToolChoice        string               `json:"toolChoice,omitempty"`
	OutputSchema      *OutputSchema        `json:"outputSchema,omitempty"`
	Temperature       *json.Number         `json:"temperature,omitempty"`
	Truncation        string               `json:"truncation,omitempty"`
	TopP              *json.Number         `json:"topP,omitempty"`
	Metadata          map[string]any       `json:"metadata,omitempty"`
	Tools             []ToolUseDefinition  `json:"tools,omitzero"`
	InputAsToolResult *bool                `json:"inputAsToolResult,omitempty"`
}

func (r CompletionRequest) Reset() CompletionRequest {
	r.Input = nil
	r.InputAsToolResult = &[]bool{false}[0]
	r.NewThread = false
	return r
}

type ToolUseDefinition struct {
	Name        string          `json:"name,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
	Description string          `json:"description,omitempty"`
	Attributes  map[string]any  `json:"-"`
}

type CompletionProgress struct {
	Model   string         `json:"model,omitempty"`
	Partial bool           `json:"partial,omitempty"`
	HasMore bool           `json:"hasMore,omitempty"`
	Item    CompletionItem `json:"item,omitempty"`
}

const CompletionProgressMetaKey = "ai.nanobot.progress/completion"

type CompletionItem struct {
	ID             string               `json:"id,omitempty"`
	Message        *mcp.SamplingMessage `json:"message,omitempty"`
	ToolCall       *ToolCall            `json:"toolCall,omitempty"`
	ToolCallResult *ToolCallResult      `json:"toolCallResult,omitempty"`
	Reasoning      *Reasoning           `json:"reasoning,omitempty"`
}

type Reasoning struct {
	ID               string        `json:"id,omitempty"`
	EncryptedContent string        `json:"encryptedContent,omitempty"`
	Summary          []SummaryText `json:"summary,omitempty"`
}

type SummaryText struct {
	Text string `json:"text,omitempty"`
}

type CompletionResponse struct {
	Output       []CompletionItem `json:"output,omitempty"`
	ChatResponse bool             `json:"chatResponse,omitempty"`
	Model        string           `json:"model,omitempty"`
}

type ToolCallResult struct {
	OutputRole string     `json:"outputRole,omitempty"`
	CallID     string     `json:"call_id,omitempty"`
	Output     CallResult `json:"output,omitempty"`
}

type ToolCall struct {
	Arguments  string `json:"arguments,omitempty"`
	CallID     string `json:"call_id,omitempty"`
	Name       string `json:"name,omitempty"`
	ID         string `json:"id,omitempty"`
	Target     string `json:"target,omitempty"`
	TargetType string `json:"targetType,omitempty"`
}

type CallResult struct {
	Content      []mcp.Content
	IsError      bool   `json:"isError,omitempty"`
	ChatResponse bool   `json:"chatResponse,omitempty"`
	Agent        string `json:"agent,omitempty"`
	Model        string `json:"model,omitempty"`
	StopReason   string `json:"stopReason,omitempty"`
}
