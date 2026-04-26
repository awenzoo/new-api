package service

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service/auto_router_rule"
	"github.com/gin-gonic/gin"
)

const AutoModelName = "AUTO"

var autoRouteRules = []auto_router_rule.Rule{
	auto_router_rule.Multimodal(),
}

func IsAutoModel(modelName string) bool {
	return strings.EqualFold(strings.TrimSpace(modelName), AutoModelName)
}

// RouteAutoModel parses the request body and returns the best model name.
// Falls back to DefaultRoutedModel on parse failure.
func RouteAutoModel(c *gin.Context) string {
	req := &auto_router_rule.Request{}
	if err := common.UnmarshalBodyReusable(c, req); err != nil {
		common.SysLog("[AUTO] parse request body failed: " + err.Error())
		return auto_router_rule.DefaultRoutedModel
	}
	if ruleName, model, ok := auto_router_rule.MatchFirst(autoRouteRules, req); ok {
		common.SysLog(fmt.Sprintf("[AUTO] rule=%s → model=%s", ruleName, model))
		c.Set("auto_matched_rule", ruleName)
		return model
	}
	common.SysLog(fmt.Sprintf("[AUTO] no rule matched → default model=%s", auto_router_rule.DefaultRoutedModel))
	return auto_router_rule.DefaultRoutedModel
}

// RouteAutoModelStatic returns the default routed model without parsing a request.
// Used by channel tests and other contexts without a gin.Context.
func RouteAutoModelStatic() string {
	return auto_router_rule.DefaultRoutedModel
}
