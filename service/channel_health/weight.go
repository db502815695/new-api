package channel_health

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

// effectiveWeightWithProbe 计算有效权重，同时返回是否为探测保底值。
func effectiveWeightWithProbe(ch *model.Channel, base int) (eff int, isProbe bool) {
	if !ch.HealthMonitorEnabled {
		return base, false
	}
	if !operation_setting.ChannelHealthEnabled() {
		return base, false
	}
	snap := GetSnapshot(ch.Id)
	minReqs := int64(operation_setting.ChannelHealthMinRequests())
	if snap.Requests < minReqs {
		return base, false
	}
	penalty := 1.0
	errThreshold := operation_setting.ChannelHealthErrorRateThreshold()
	if snap.ErrorRate > errThreshold {
		penalty *= (1 - snap.ErrorRate)
	}
	ttftThreshold := operation_setting.ChannelHealthTtftThresholdMs()
	if snap.AvgTtftMs > ttftThreshold && snap.AvgTtftMs > 0 {
		penalty *= float64(ttftThreshold) / float64(snap.AvgTtftMs)
	}
	if penalty < 0 {
		penalty = 0
	}
	computed := int(float64(base) * penalty)

	// 探测保底：baseWeight>0 时至少保留 5% 流量用于探测恢复
	if base > 0 {
		minProbe := int(float64(base) * 0.05)
		if minProbe < 1 {
			minProbe = 1
		}
		if computed < minProbe {
			return minProbe, true
		}
		if computed < 1 {
			computed = 1
		}
	}
	return computed, false
}

// EffectiveWeight 计算渠道有效权重（原权重 × 健康惩罚系数）
func EffectiveWeight(ch *model.Channel) int {
	eff, _ := effectiveWeightWithProbe(ch, ch.GetWeight())
	return eff
}

// EffectiveWeightWithProbeFlag 同 EffectiveWeight，额外返回是否为探测保底权重。
func EffectiveWeightWithProbeFlag(ch *model.Channel, base int) (int, bool) {
	return effectiveWeightWithProbe(ch, base)
}

func init() {
	model.EffectiveWeightFunc = EffectiveWeight
	model.ExcludeChannelFunc = func(ch *model.Channel) bool {
		if ch == nil {
			return true
		}
		if ch.Status != common.ChannelStatusEnabled {
			return true
		}
		if !ch.HealthMonitorEnabled {
			return false
		}
		return IsUnhealthy(ch.Id)
	}
}
