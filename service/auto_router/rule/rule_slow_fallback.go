package rule

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

type slowFallbackRule struct{}

// SlowFallback 当 GLM-5.1 近期首字响应慢时，回退到 GLM-5-Turbo
func SlowFallback() Rule {
	return slowFallbackRule{}
}

func (slowFallbackRule) Name() string { return "slow_fallback" }

func (slowFallbackRule) TargetModel() string { return SlowFallbackRoutedModel }

var (
	frtCacheMu    sync.RWMutex
	frtCacheVal   float64
	frtCacheTime  time.Time
	frtCacheTTL   = 60 * time.Second
)

func (slowFallbackRule) Match(_ *Request) bool {
	if model.LOG_DB == nil {
		return false
	}

	avgFrt := getCachedAvgFrt()
	if avgFrt == 0 {
		return false
	}

	return avgFrt > float64(common.GetEnvOrDefault("AUTO_SLOW_THRESHOLD_MS", SlowThresholdFirstTokenMs))
}

func getCachedAvgFrt() float64 {
	frtCacheMu.RLock()
	if time.Since(frtCacheTime) < frtCacheTTL {
		val := frtCacheVal
		frtCacheMu.RUnlock()
		return val
	}
	frtCacheMu.RUnlock()

	frtCacheMu.Lock()
	defer frtCacheMu.Unlock()

	if time.Since(frtCacheTime) < frtCacheTTL {
		return frtCacheVal
	}

	frtCacheVal = queryAvgFrt()
	frtCacheTime = time.Now()
	return frtCacheVal
}

func queryAvgFrt() float64 {
	since := time.Now().Add(-time.Duration(slowCheckWindow) * time.Minute).Unix()

	var frtExpr string
	var frtGtZero string
	if common.UsingPostgreSQL {
		frtExpr = "(other::json->>'frt')::real"
		frtGtZero = "(other::json->>'frt')::real > 0"
	} else {
		frtExpr = "CAST(json_extract(other, '$.frt') AS REAL)"
		frtGtZero = "json_extract(other, '$.frt') > 0"
	}

	var avgFrt float64
	err := model.LOG_DB.Model(&model.Log{}).
		Where("model_name = ? AND created_at >= ? AND is_stream = ? AND "+frtGtZero, DefaultRoutedModel, since, true).
		Select(fmt.Sprintf("COALESCE(AVG(%s), 0)", frtExpr)).
		Row().Scan(&avgFrt)
	if err != nil {
		return 0
	}
	return avgFrt
}
