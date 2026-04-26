package service

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

const AutoModelName = "AUTO"

const DefaultAutoRoutedModel = "GML-5.1"
const MultimodalRoutedModel = "qwen3.6-plus"

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
	Messages []autoRouteMessage `json:"messages,omitempty"`
}

type autoRouteMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

func routeRequest(req *autoRouteRequest) string {
	if hasImage(req.Messages) {
		return MultimodalRoutedModel
	}
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
