package rule

type multimodalRule struct{}

// Multimodal 匹配包含图片的请求（同时支持 OpenAI 和 Anthropic 格式）
func Multimodal() Rule {
	return multimodalRule{}
}

func (multimodalRule) Name() string { return "multimodal" }

func (multimodalRule) Match(req *Request) bool {
	for _, msg := range req.Messages {
		parts, ok := msg.Content.([]interface{})
		if !ok {
			continue
		}
		for _, item := range parts {
			part, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			// "image_url": OpenAI 格式, "image": Anthropic 格式
			t, _ := part["type"].(string)
			if t == "image_url" || t == "image" {
				return true
			}
		}
	}
	return false
}

func (multimodalRule) TargetModel() string { return MultimodalRoutedModel }
