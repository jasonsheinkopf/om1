package llm

// ============================================================================
// STUDY GUIDE — LLM CONTRACTS / TYPED CORTEX OUTPUT
//
// This file defines the abstraction between OM1 and concrete LLM providers.
//
// CRITICAL IDEA:
//   Cortex does NOT merely emit prose that OM1 guesses commands from.
//   Response can contain TextContent plus structured []ToolCall.
//   ToolCall has a Name and structured Arguments.
//
// Action schemas tell the model which callable actions exist and what arguments
// they expect. Concrete Gemini/OpenAI/etc. plugins satisfy the LLM interface.
// ============================================================================

import (
	"context"
)

type Message struct {
	Role    string // "system" | "user" | "assistant"
	Content string
}

// ToolCall is the symbolic instruction returned by the model before OM1 resolves
// it to a concrete AgentAction/Connector.
type ToolCall struct {
	Name      string
	Arguments map[string]any
}

// Response is Cortex's typed result: natural-language content, structured tool calls,
// and usage metadata can travel together.
type Response struct {
	TextContent string
	ToolCalls   []ToolCall
	Usage       Usage
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
}

// LLM is the provider contract. Runtime code depends on this interface rather than
// hard-coding Gemini/OpenAI/etc., which keeps the architecture pluggable.
type LLM interface {
	// Call sends a prompt and conversation history to the model.
	Call(ctx context.Context, prompt string, history []Message) (*Response, error)

	// SetSchemas configures the tool schemas that the LLM can use for tool calls.
	SetSchemas(schemas []map[string]any)

	// FunctionSchemas returns the currently configured tool schemas.
	FunctionSchemas() []map[string]any
}

// Factory is a function type for creating LLM instances with a given configuration.
type Factory func(cfg map[string]any) (LLM, error)

// registry holds the mapping of LLM type names to their corresponding factory functions.
var registry = map[string]Factory{}

// Register adds a new LLM factory to the registry under the specified type name.
func Register(typeName string, f Factory) {
	registry[typeName] = f
}

// Load creates an LLM instance based on the provided type name and configuration.
func Load(typeName string, cfg map[string]any) (LLM, error) {
	f, ok := registry[typeName]
	if !ok {
		return nil, &UnknownPluginError{Name: typeName}
	}
	return f(cfg)
}

// UnknownPluginError is returned when an attempt is made to load an LLM plugin that is not registered.
type UnknownPluginError struct{ Name string }

// Error returns a descriptive error message indicating that the specified LLM plugin was not found.
func (e *UnknownPluginError) Error() string { return "llm plugin not found: " + e.Name }

// SpeakText returns the text from the speak action.
func (r *Response) SpeakText() string {
	for _, tc := range r.ToolCalls {
		if tc.Name == "speak" {
			if text, ok := tc.Arguments["action"].(string); ok {
				return text
			}
		}
	}
	return r.TextContent
}
