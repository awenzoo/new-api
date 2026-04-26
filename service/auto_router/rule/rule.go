package rule

// Rule 判断请求是否应路由到特定模型
type Rule interface {
	Name() string
	Match(req *Request) bool
	TargetModel() string
}

type Request struct {
	Messages []Message `json:"messages,omitempty"`
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// MatchFirst 遍历规则列表，返回首个匹配的规则
func MatchFirst(rules []Rule, req *Request) (ruleName string, targetModel string, matched bool) {
	for _, rule := range rules {
		if rule.Match(req) {
			return rule.Name(), rule.TargetModel(), true
		}
	}
	return "", "", false
}
