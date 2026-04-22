package service

import (
	"testing"
)

func TestIsAutoModel(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"AUTO", true},
		{"auto", true},
		{"Auto", true},
		{" AUTO ", true},
		{"auto ", true},
		{"qwen3.6-plus", false},
		{"glm-5.1", false},
		{"", false},
		{"AUTOMODEL", false},
	}
	for _, tt := range tests {
		if got := IsAutoModel(tt.input); got != tt.want {
			t.Errorf("IsAutoModel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func TestRouteRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *autoRouteRequest
		want string
	}{
		{
			name: "L0: max_tokens > 16384 routes to glm-5.1",
			req:  &autoRouteRequest{MaxTokens: intPtr(32768)},
			want: "glm-5.1",
		},
		{
			name: "L0: max_tokens = 16384 does not trigger glm-5.1",
			req:  &autoRouteRequest{MaxTokens: intPtr(16384)},
			want: "qwen3.5-plus",
		},
		{
			name: "L0: max_tokens nil falls through",
			req:  &autoRouteRequest{},
			want: "qwen3.5-plus",
		},
		{
			name: "L0: image in messages routes to qwen3.6-plus",
			req: &autoRouteRequest{
				Messages: []autoRouteMessage{
					{
						Role: "user",
						Content: []interface{}{
							map[string]interface{}{"type": "text", "text": "describe this"},
							map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "http://example.com/img.png"}},
						},
					},
				},
			},
			want: "qwen3.6-plus",
		},
		{
			name: "L0: max_tokens takes priority over image",
			req: &autoRouteRequest{
				MaxTokens: intPtr(20000),
				Messages: []autoRouteMessage{
					{
						Role: "user",
						Content: []interface{}{
							map[string]interface{}{"type": "image_url"},
						},
					},
				},
			},
			want: "glm-5.1",
		},
		{
			name: "L1: thinking routes to glm-5.1",
			req:  &autoRouteRequest{Thinking: map[string]interface{}{"type": "enabled"}},
			want: "glm-5.1",
		},
		{
			name: "L1: budget_tokens routes to glm-5.1",
			req:  &autoRouteRequest{BudgetTokens: intPtr(10000)},
			want: "glm-5.1",
		},
		{
			name: "L1: reasoning_effort routes to glm-5.1",
			req:  &autoRouteRequest{ReasoningEffort: strPtr("high")},
			want: "glm-5.1",
		},
		{
			name: "L1: tools routes to qwen3.6-plus",
			req:  &autoRouteRequest{Tools: []interface{}{map[string]interface{}{"type": "function"}}},
			want: "qwen3.6-plus",
		},
		{
			name: "L1: functions routes to qwen3.6-plus",
			req:  &autoRouteRequest{Functions: []interface{}{map[string]interface{}{"name": "get_weather"}}},
			want: "qwen3.6-plus",
		},
		{
			name: "L1: code keyword routes to qwen3.6-plus",
			req: &autoRouteRequest{
				Messages: []autoRouteMessage{
					{Role: "user", Content: "写一段代码实现快速排序"},
				},
			},
			want: "qwen3.6-plus",
		},
		{
			name: "L2: plain conversation defaults to qwen3.5-plus",
			req: &autoRouteRequest{
				Messages: []autoRouteMessage{
					{Role: "user", Content: "你好，今天天气怎么样？"},
				},
			},
			want: "qwen3.5-plus",
		},
		{
			name: "L2: empty request defaults to qwen3.5-plus",
			req:  &autoRouteRequest{},
			want: "qwen3.5-plus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routeRequest(tt.req)
			if got != tt.want {
				t.Errorf("routeRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasImage(t *testing.T) {
	tests := []struct {
		name     string
		messages []autoRouteMessage
		want     bool
	}{
		{
			name:     "nil messages",
			messages: nil,
			want:     false,
		},
		{
			name:     "string content",
			messages: []autoRouteMessage{{Role: "user", Content: "hello"}},
			want:     false,
		},
		{
			name: "array content with image",
			messages: []autoRouteMessage{
				{
					Role: "user",
					Content: []interface{}{
						map[string]interface{}{"type": "image_url", "image_url": "http://example.com/img.png"},
					},
				},
			},
			want: true,
		},
		{
			name: "array content without image",
			messages: []autoRouteMessage{
				{
					Role: "user",
					Content: []interface{}{
						map[string]interface{}{"type": "text", "text": "hello"},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasImage(tt.messages); got != tt.want {
				t.Errorf("hasImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasCodeSignals(t *testing.T) {
	tests := []struct {
		name     string
		messages []autoRouteMessage
		want     bool
	}{
		{
			name:     "nil messages",
			messages: nil,
			want:     false,
		},
		{
			name:     "no code keywords",
			messages: []autoRouteMessage{{Role: "user", Content: "你好"}},
			want:     false,
		},
		{
			name:     "code keyword in string content",
			messages: []autoRouteMessage{{Role: "user", Content: "写一段代码实现快速排序"}},
			want:     true,
		},
		{
			name: "code keyword in array content text field",
			messages: []autoRouteMessage{
				{
					Role: "user",
					Content: []interface{}{
						map[string]interface{}{"type": "text", "text": "please debug this error"},
					},
				},
			},
			want: true,
		},
		{
			name:     "case insensitive matching",
			messages: []autoRouteMessage{{Role: "user", Content: "I need to CODE something"}},
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasCodeSignals(tt.messages); got != tt.want {
				t.Errorf("hasCodeSignals() = %v, want %v", got, tt.want)
			}
		})
	}
}
