package rule

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
)

type advancedRule struct{}

// Advanced 匹配最后一条用户消息中包含代码审查/分析/重构等关键词的请求
func Advanced() Rule {
	return advancedRule{}
}

func (advancedRule) Name() string { return "advanced" }

func (advancedRule) TargetModel() string { return AdvancedRoutedModel }

var advancedKeywords = []string{
	"代码审查",
	"代码评审",
	"代码走查",
	"代码检查",
	"代码分析",
	"代码审计",
	"代码重构",
	"代码优化",
	"审查代码",
	"评审代码",
	"检查代码",
	"分析代码",
	"重构代码",
	"优化代码",
	"code review",
	"code audit",
	"代码review",
	"review代码",
	"代码audit",
	"audit代码",
}

func (advancedRule) Match(req *Request) bool {
	if !common.GetEnvOrDefaultBool("AUTO_ADVANCED_ENABLED", true) {
		return false
	}
	lastUserMsg := lastUserText(req)
	if lastUserMsg == "" {
		return false
	}
	lower := strings.ToLower(lastUserMsg)
	for _, kw := range advancedKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// lastUserText 提取最后一条 role=user 的消息文本内容
func lastUserText(req *Request) string {
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role != "user" {
			continue
		}
		switch v := req.Messages[i].Content.(type) {
		case string:
			return v
		case []interface{}:
			for _, item := range v {
				part, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				t, _ := part["type"].(string)
				if t == "text" {
					if s, ok := part["text"].(string); ok {
						return s
					}
				}
			}
		}
		return ""
	}
	return ""
}
