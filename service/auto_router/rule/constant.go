package rule

const (
	DefaultRoutedModel       = "GLM-5.1"
	MultimodalRoutedModel    = "qwen3.6-plus"
	SlowFallbackRoutedModel  = "GLM-5-Turbo"
	AdvancedRoutedModel      = "anthropic/claude-opus-4.6"

	// SlowThresholdFirstTokenMs 首字响应时间超过此值（毫秒）则判定为慢
	SlowThresholdFirstTokenMs = 10000
	slowCheckWindow           = 10 // 分钟
)
