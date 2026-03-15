package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	forcedOpenAIUpstreamModel           = "gpt-5.2-codex"
	forcedOpenAIUpstreamReasoningEffort = "xhigh"
)

func forcedOpenAIReasoningEffortPtr() *string {
	effort := forcedOpenAIUpstreamReasoningEffort
	return &effort
}

func normalizeOpenAIUpstreamMessageContentItem(text string) map[string]any {
	return map[string]any{
		"type": "input_text",
		"text": text,
	}
}

func normalizeOpenAIUpstreamMessageItem(content []any) map[string]any {
	return map[string]any{
		"type":    "message",
		"role":    "user",
		"content": content,
	}
}

func normalizeOpenAIUpstreamContentItem(item any) (map[string]any, bool) {
	itemMap, ok := item.(map[string]any)
	if !ok || len(itemMap) == 0 {
		return nil, false
	}
	itemType, _ := itemMap["type"].(string)
	itemType = strings.TrimSpace(itemType)
	if itemType != "text" && itemType != "input_text" {
		return nil, false
	}
	text, _ := itemMap["text"].(string)
	if itemType == "input_text" {
		return itemMap, true
	}
	return normalizeOpenAIUpstreamMessageContentItem(text), true
}

func normalizeOpenAIUpstreamInputMap(reqBody map[string]any) bool {
	if len(reqBody) == 0 {
		return false
	}

	rawInput, exists := reqBody["input"]
	if !exists || rawInput == nil {
		return false
	}

	switch input := rawInput.(type) {
	case string:
		reqBody["input"] = []any{normalizeOpenAIUpstreamMessageItem([]any{normalizeOpenAIUpstreamMessageContentItem(input)})}
		return true
	case map[string]any:
		itemType, _ := input["type"].(string)
		if strings.TrimSpace(itemType) == "message" {
			reqBody["input"] = []any{input}
			return true
		}
		if normalizedContent, ok := normalizeOpenAIUpstreamContentItem(input); ok {
			reqBody["input"] = []any{normalizeOpenAIUpstreamMessageItem([]any{normalizedContent})}
			return true
		}
		reqBody["input"] = []any{input}
		return true
	case []any:
		if len(input) == 0 {
			return false
		}
		contentItems := make([]any, 0, len(input))
		for _, item := range input {
			normalizedContent, ok := normalizeOpenAIUpstreamContentItem(item)
			if !ok {
				return false
			}
			contentItems = append(contentItems, normalizedContent)
		}
		reqBody["input"] = []any{normalizeOpenAIUpstreamMessageItem(contentItems)}
		return true
	default:
		return false
	}
}

func normalizeOpenAIUpstreamInputBytes(body []byte) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}
	if !gjson.GetBytes(body, "input").Exists() {
		return body, false, nil
	}

	var reqBody map[string]any
	if err := json.Unmarshal(body, &reqBody); err != nil {
		return body, false, fmt.Errorf("unmarshal request body for input normalization: %w", err)
	}
	if !normalizeOpenAIUpstreamInputMap(reqBody) {
		return body, false, nil
	}
	normalized, err := json.Marshal(reqBody)
	if err != nil {
		return body, false, fmt.Errorf("marshal normalized input request body: %w", err)
	}
	return normalized, true, nil
}

func enforceOpenAIUpstreamRequestMap(reqBody map[string]any) bool {
	if len(reqBody) == 0 {
		return false
	}

	changed := false
	if model, _ := reqBody["model"].(string); strings.TrimSpace(model) != forcedOpenAIUpstreamModel {
		reqBody["model"] = forcedOpenAIUpstreamModel
		changed = true
	}

	reasoning, ok := reqBody["reasoning"].(map[string]any)
	if !ok || reasoning == nil {
		reasoning = make(map[string]any)
		reqBody["reasoning"] = reasoning
		changed = true
	}
	if effort, _ := reasoning["effort"].(string); strings.TrimSpace(effort) != forcedOpenAIUpstreamReasoningEffort {
		reasoning["effort"] = forcedOpenAIUpstreamReasoningEffort
		changed = true
	}

	if _, exists := reqBody["reasoning_effort"]; exists {
		delete(reqBody, "reasoning_effort")
		changed = true
	}

	return changed
}

func enforceOpenAIUpstreamRequestBytes(body []byte) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}

	normalized := body
	changed := false

	if strings.TrimSpace(gjson.GetBytes(normalized, "model").String()) != forcedOpenAIUpstreamModel {
		next, err := sjson.SetBytes(normalized, "model", forcedOpenAIUpstreamModel)
		if err != nil {
			return body, false, fmt.Errorf("enforce model in request body: %w", err)
		}
		normalized = next
		changed = true
	}

	if strings.TrimSpace(gjson.GetBytes(normalized, "reasoning.effort").String()) != forcedOpenAIUpstreamReasoningEffort {
		next, err := sjson.SetBytes(normalized, "reasoning.effort", forcedOpenAIUpstreamReasoningEffort)
		if err != nil {
			return body, false, fmt.Errorf("enforce reasoning.effort in request body: %w", err)
		}
		normalized = next
		changed = true
	}

	if gjson.GetBytes(normalized, "reasoning_effort").Exists() {
		next, err := sjson.DeleteBytes(normalized, "reasoning_effort")
		if err != nil {
			return body, false, fmt.Errorf("delete reasoning_effort in request body: %w", err)
		}
		normalized = next
		changed = true
	}

	return normalized, changed, nil
}
