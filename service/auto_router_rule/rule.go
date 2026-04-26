package auto_router_rule

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

func MatchFirst(rules []Rule, req *Request) (ruleName string, targetModel string, matched bool) {
	for _, rule := range rules {
		if rule.Match(req) {
			return rule.Name(), rule.TargetModel(), true
		}
	}
	return "", "", false
}
