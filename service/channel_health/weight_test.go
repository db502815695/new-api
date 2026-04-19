package channel_health

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

func enableHealthSetting() {
	s := operation_setting.GetChannelHealthSetting()
	s.Enabled = true
	s.MinRequests = 5
	s.ErrorRateThreshold = 0.3
	s.TtftThresholdMs = 5000
}

func disableHealthSetting() {
	operation_setting.GetChannelHealthSetting().Enabled = false
}

func newMockChannel(id int, weight uint, monitorEnabled bool) *model.Channel {
	w := weight
	return &model.Channel{
		Id:                   id,
		Weight:               &w,
		HealthMonitorEnabled: monitorEnabled,
	}
}

// seedTracker 向 trackers 中注入足量样本以满足 MinRequests。
func seedTracker(id int, reqs, errs int64) {
	tr := newTracker()
	// 直接写桶 0 当前时间 —— 手动绕过 rotate 使用 Record 方式更稳健
	for i := int64(0); i < reqs; i++ {
		isErr := i < errs
		tr.Record(isErr, 100)
	}
	trackers.Store(id, tr)
}

func cleanTracker(id int) {
	trackers.Delete(id)
}

func TestEffectiveWeightNoMonitor(t *testing.T) {
	enableHealthSetting()
	defer disableHealthSetting()
	ch := newMockChannel(99001, 100, false)
	eff, probe := effectiveWeightWithProbe(ch, 100)
	if eff != 100 || probe {
		t.Fatalf("unmonitored channel: want eff=100,probe=false; got eff=%d,probe=%v", eff, probe)
	}
}

func TestEffectiveWeightHealthSwitchOff(t *testing.T) {
	// 全局开关关闭时直接返回 base
	disableHealthSetting()
	ch := newMockChannel(99002, 80, true)
	eff, probe := effectiveWeightWithProbe(ch, 80)
	if eff != 80 || probe {
		t.Fatalf("switch off: want eff=80,probe=false; got eff=%d,probe=%v", eff, probe)
	}
}

func TestEffectiveWeightProbeFloor(t *testing.T) {
	enableHealthSetting()
	defer func() {
		disableHealthSetting()
		cleanTracker(99003)
	}()
	ch := newMockChannel(99003, 100, true)
	// 10 请求全部失败 => errorRate=1.0 >> 阈值 0.3
	seedTracker(99003, 10, 10)

	eff, probe := effectiveWeightWithProbe(ch, 100)
	if !probe {
		t.Fatalf("expected probe=true for 100%% error rate, got eff=%d", eff)
	}
	// 探测保底 = max(1, int(100*0.05)) = 5
	if eff != 5 {
		t.Fatalf("expected probe floor=5, got %d", eff)
	}
}

func TestEffectiveWeightProbeFloorZeroBase(t *testing.T) {
	enableHealthSetting()
	defer func() {
		disableHealthSetting()
		cleanTracker(99004)
	}()
	ch := newMockChannel(99004, 0, true)
	seedTracker(99004, 10, 10)

	eff, probe := effectiveWeightWithProbe(ch, 0)
	if probe || eff != 0 {
		t.Fatalf("base=0: want eff=0,probe=false; got eff=%d,probe=%v", eff, probe)
	}
}

func TestEffectiveWeightHealthy(t *testing.T) {
	enableHealthSetting()
	defer func() {
		disableHealthSetting()
		cleanTracker(99005)
	}()
	ch := newMockChannel(99005, 100, true)
	// 10 请求 0 错误 => 无惩罚
	seedTracker(99005, 10, 0)

	eff, probe := effectiveWeightWithProbe(ch, 100)
	if probe {
		t.Fatal("healthy channel: expected probe=false")
	}
	if eff != 100 {
		t.Fatalf("healthy channel: expected eff=100, got %d", eff)
	}
}

func TestEffectiveWeightMinimumOneForBase1(t *testing.T) {
	enableHealthSetting()
	defer func() {
		disableHealthSetting()
		cleanTracker(99006)
	}()
	ch := newMockChannel(99006, 1, true)
	// 10 请求全部失败
	seedTracker(99006, 10, 10)

	eff, probe := effectiveWeightWithProbe(ch, 1)
	if !probe {
		t.Fatal("base=1 full failure: expected probe=true")
	}
	// int(1*0.05)=0, 向上取 1
	if eff < 1 {
		t.Fatalf("expected eff>=1, got %d", eff)
	}
}
