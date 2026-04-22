package service

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

const AutoModelName = "AUTO"

const DefaultAutoRoutedModel = "qwen3.5-plus"

func IsAutoModel(modelName string) bool {
	return strings.EqualFold(strings.TrimSpace(modelName), AutoModelName)
}

// RouteAutoModel parses the request body and returns the best model name.
// Falls back to DefaultAutoRoutedModel on parse failure.
func RouteAutoModel(c *gin.Context) string {
	req := &autoRouteRequest{}
	if err := common.UnmarshalBodyReusable(c, req); err != nil {
		return DefaultAutoRoutedModel
	}
	return routeRequest(req)
}

// --- internal types and logic ---

type autoRouteRequest struct {
	MaxTokens       *int               `json:"max_tokens,omitempty"`
	Messages        []autoRouteMessage `json:"messages,omitempty"`
	Tools           []interface{}      `json:"tools,omitempty"`
	Functions       []interface{}      `json:"functions,omitempty"`
	Thinking        interface{}        `json:"thinking,omitempty"`
	BudgetTokens    *int               `json:"budget_tokens,omitempty"`
	ReasoningEffort *string            `json:"reasoning_effort,omitempty"`
}

type autoRouteMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

func routeRequest(req *autoRouteRequest) string {
	// L0: hard constraints
	if req.MaxTokens != nil && *req.MaxTokens > 16384 {
		return "glm-5.1"
	}
	if hasImage(req.Messages) {
		return "qwen3.6-plus"
	}

	// L1: upgrade signals
	if req.Thinking != nil || req.BudgetTokens != nil || (req.ReasoningEffort != nil && *req.ReasoningEffort != "") {
		return "glm-5.1"
	}
	if len(req.Tools) > 0 || len(req.Functions) > 0 {
		return "qwen3.6-plus"
	}
	if hasCodeSignals(req.Messages) {
		return "qwen3.6-plus"
	}

	// L2: default
	return DefaultAutoRoutedModel
}

func hasImage(messages []autoRouteMessage) bool {
	for _, msg := range messages {
		parts, ok := msg.Content.([]interface{})
		if !ok {
			continue
		}
		for _, item := range parts {
			part, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if part["type"] == "image_url" {
				return true
			}
		}
	}
	return false
}

var codeKeywords = []string{
	"code", "编程", "debug", "函数", "implement",
	"refactor", "代码", "coding", "programming",
}

func hasCodeSignals(messages []autoRouteMessage) bool {
	var text string
	for _, msg := range messages {
		switch content := msg.Content.(type) {
		case string:
			text += " " + strings.ToLower(content)
		case []interface{}:
			for _, item := range content {
				part, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				if part["type"] == "text" {
					if s, ok := part["text"].(string); ok {
						text += " " + strings.ToLower(s)
					}
				}
			}
		}
	}
	for _, kw := range codeKeywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
