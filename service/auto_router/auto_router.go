package auto_router

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service/auto_router/rule"
	"github.com/gin-gonic/gin"
)

// AutoModelName 触发自动路由的虚拟模型名称
const AutoModelName = "AUTO"

// autoRouteRules 按优先级排列的路由规则列表，首个匹配生效
var autoRouteRules = []rule.Rule{
	rule.Multimodal(),
	rule.SlowFallback(),
}

// IsAutoModel 判断模型名称是否为 AUTO
func IsAutoModel(modelName string) bool {
	return strings.EqualFold(strings.TrimSpace(modelName), AutoModelName)
}

// RouteAutoModel 解析请求体并返回最合适的模型名称。
// 解析失败或无规则匹配时回退到 DefaultRoutedModel。
func RouteAutoModel(c *gin.Context) string {
	req := &rule.Request{}
	if err := common.UnmarshalBodyReusable(c, req); err != nil {
		common.SysLog("[AUTO] parse request body failed: " + err.Error())
		return rule.DefaultRoutedModel
	}
	if ruleName, model, ok := rule.MatchFirst(autoRouteRules, req); ok {
		common.SysLog(fmt.Sprintf("[AUTO] rule=%s → model=%s", ruleName, model))
		c.Set("auto_matched_rule", ruleName)
		return model
	}
	common.SysLog(fmt.Sprintf("[AUTO] no rule matched → default model=%s", rule.DefaultRoutedModel))
	return rule.DefaultRoutedModel
}

// RouteAutoModelStatic 不解析请求，直接返回默认路由模型。
// 用于渠道测试等无 gin.Context 的场景。
func RouteAutoModelStatic() string {
	return rule.DefaultRoutedModel
}
