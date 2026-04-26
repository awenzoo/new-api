package service

import (
	"testing"

	"github.com/QuantumNous/new-api/service/auto_router_rule"
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
		{"GLM-5.1", false},
		{"", false},
		{"AUTOMODEL", false},
	}
	for _, tt := range tests {
		if got := IsAutoModel(tt.input); got != tt.want {
			t.Errorf("IsAutoModel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRouteRules(t *testing.T) {
	tests := []struct {
		name string
		req  *auto_router_rule.Request
		want string
	}{
		{
			name: "image in messages routes to qwen3.6-plus",
			req: &auto_router_rule.Request{
				Messages: []auto_router_rule.Message{
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
			name: "plain conversation defaults to GLM-5.1",
			req: &auto_router_rule.Request{
				Messages: []auto_router_rule.Message{
					{Role: "user", Content: "你好，今天天气怎么样？"},
				},
			},
			want: "GLM-5.1",
		},
		{
			name: "empty request defaults to GLM-5.1",
			req:  &auto_router_rule.Request{},
			want: "GLM-5.1",
		},
		{
			name: "code keyword still defaults to GLM-5.1",
			req: &auto_router_rule.Request{
				Messages: []auto_router_rule.Message{
					{Role: "user", Content: "写一段代码实现快速排序"},
				},
			},
			want: "GLM-5.1",
		},
		{
			name: "long message defaults to GLM-5.1",
			req: &auto_router_rule.Request{
				Messages: []auto_router_rule.Message{
					{Role: "user", Content: "帮我查一下天气"},
				},
			},
			want: "GLM-5.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, model, ok := auto_router_rule.MatchFirst(autoRouteRules, tt.req)
			var got string
			if ok {
				got = model
			} else {
				got = auto_router_rule.DefaultRoutedModel
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMultimodalRule(t *testing.T) {
	tests := []struct {
		name     string
		messages []auto_router_rule.Message
		want     bool
	}{
		{
			name:     "nil messages",
			messages: nil,
			want:     false,
		},
		{
			name:     "string content",
			messages: []auto_router_rule.Message{{Role: "user", Content: "hello"}},
			want:     false,
		},
		{
			name: "array content with image",
			messages: []auto_router_rule.Message{
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
			messages: []auto_router_rule.Message{
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
	rule := auto_router_rule.Multimodal()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &auto_router_rule.Request{Messages: tt.messages}
			if got := rule.Match(req); got != tt.want {
				t.Errorf("Multimodal().Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
