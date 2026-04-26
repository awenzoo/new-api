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
		{"GML-5.1", false},
		{"", false},
		{"AUTOMODEL", false},
	}
	for _, tt := range tests {
		if got := IsAutoModel(tt.input); got != tt.want {
			t.Errorf("IsAutoModel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRouteRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *autoRouteRequest
		want string
	}{
		{
			name: "image in messages routes to qwen3.6-plus",
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
			name: "plain conversation defaults to GML-5.1",
			req: &autoRouteRequest{
				Messages: []autoRouteMessage{
					{Role: "user", Content: "你好，今天天气怎么样？"},
				},
			},
			want: "GML-5.1",
		},
		{
			name: "empty request defaults to GML-5.1",
			req:  &autoRouteRequest{},
			want: "GML-5.1",
		},
		{
			name: "code keyword still defaults to GML-5.1",
			req: &autoRouteRequest{
				Messages: []autoRouteMessage{
					{Role: "user", Content: "写一段代码实现快速排序"},
				},
			},
			want: "GML-5.1",
		},
		{
			name: "long message defaults to GML-5.1",
			req: &autoRouteRequest{
				Messages: []autoRouteMessage{
					{Role: "user", Content: "帮我查一下天气"},
				},
			},
			want: "GML-5.1",
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
