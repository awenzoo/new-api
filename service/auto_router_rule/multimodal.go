package auto_router_rule

type multimodalRule struct{}

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
			if part["type"] == "image_url" {
				return true
			}
		}
	}
	return false
}

func (multimodalRule) TargetModel() string { return MultimodalRoutedModel }
