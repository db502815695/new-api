package controller

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service/channel_health"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

// channelHealthSnapshotDTO 快照响应字段
type channelHealthSnapshotDTO struct {
	ChannelId       int     `json:"channel_id"`
	ChannelName     string  `json:"channel_name"`
	MonitorEnabled  bool    `json:"monitor_enabled"`
	Priority        int64   `json:"priority"`
	BaseWeight      int     `json:"base_weight"`
	EffectiveWeight int     `json:"effective_weight"`
	IsProbeWeight   bool    `json:"is_probe_weight"`
	Weight          int     `json:"weight"`
	ErrorRate       float64 `json:"error_rate"`
	AvgTtftMs       int64   `json:"avg_ttft_ms"`
	Requests        int64   `json:"requests"`
	Errors          int64   `json:"errors"`
	WindowSeconds   int     `json:"window_seconds"`
	PenaltyReason   string  `json:"penalty_reason"`
}

const channelWeightAdjustStep = 10

func buildSnapshotDTO(ch *model.Channel, priority int64, base int) channelHealthSnapshotDTO {
	snap := channel_health.GetSnapshot(ch.Id)
	eff := base
	isProbe := false
	if ch.HealthMonitorEnabled {
		if base != ch.GetWeight() {
			cloned := *ch
			weight := uint(0)
			if base > 0 {
				weight = uint(base)
			}
			cloned.Weight = &weight
			eff, isProbe = channel_health.EffectiveWeightWithProbeFlag(&cloned, base)
		} else {
			eff, isProbe = channel_health.EffectiveWeightWithProbeFlag(ch, base)
		}
	}
	reason := ""
	errThreshold := operation_setting.ChannelHealthErrorRateThreshold()
	ttftThreshold := operation_setting.ChannelHealthTtftThresholdMs()
	if ch.HealthMonitorEnabled && snap.Requests >= int64(operation_setting.ChannelHealthMinRequests()) {
		if snap.ErrorRate > errThreshold && snap.AvgTtftMs > ttftThreshold {
			reason = "error_rate>threshold && ttft>threshold"
		} else if snap.ErrorRate > errThreshold {
			reason = "error_rate>threshold"
		} else if snap.AvgTtftMs > ttftThreshold {
			reason = "ttft>threshold"
		}
	}
	return channelHealthSnapshotDTO{
		ChannelId:       ch.Id,
		ChannelName:     ch.Name,
		MonitorEnabled:  ch.HealthMonitorEnabled,
		Priority:        priority,
		BaseWeight:      base,
		EffectiveWeight: eff,
		IsProbeWeight:   isProbe,
		Weight:          base,
		ErrorRate:       snap.ErrorRate,
		AvgTtftMs:       snap.AvgTtftMs,
		Requests:        snap.Requests,
		Errors:          snap.Errors,
		WindowSeconds:   snap.WindowSeconds,
		PenaltyReason:   reason,
	}
}

func snapshotChannelIDs(group string) []int {
	group = strings.TrimSpace(group)
	groupChannels := model.ListGroupChannels()
	seen := make(map[int]bool)
	result := make([]int, 0)
	if group != "" {
		for _, id := range groupChannels[group] {
			if seen[id] {
				continue
			}
			seen[id] = true
			result = append(result, id)
		}
		sort.Ints(result)
		return result
	}
	for _, ids := range groupChannels {
		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true
			result = append(result, id)
		}
	}
	sort.Ints(result)
	return result
}

type channelWeightAdjustRequest struct {
	Action string `json:"action"`
}

type resetGroupChannelHealthRequest struct {
	Group string `json:"group"`
}

type updateGroupAbilityRequest struct {
	Group     string `json:"group"`
	ChannelId int    `json:"channel_id"`
	Priority  *int64 `json:"priority,omitempty"`
	Weight    *uint  `json:"weight,omitempty"`
}

func loadGroupAbilityMap(group string) (map[int]model.Ability, error) {
	group = strings.TrimSpace(group)
	if group == "" {
		return nil, nil
	}
	abilities, _, err := model.GetSingleModelAbilitiesByGroup(group)
	if err != nil {
		return nil, err
	}
	result := make(map[int]model.Ability, len(abilities))
	for _, ability := range abilities {
		result[ability.ChannelId] = ability
	}
	return result, nil
}

// GetAllChannelHealthSnapshots 返回所有渠道的快照（管理员）
func GetAllChannelHealthSnapshots(c *gin.Context) {
	abilityMap, err := loadGroupAbilityMap(c.Query("group"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	result := make([]channelHealthSnapshotDTO, 0)
	for _, id := range snapshotChannelIDs(c.Query("group")) {
		ch, err := model.CacheGetChannel(id)
		if err != nil || ch == nil {
			continue
		}
		priority := ch.GetPriority()
		baseWeight := ch.GetWeight()
		if ability, ok := abilityMap[id]; ok {
			if ability.Priority != nil {
				priority = *ability.Priority
			}
			baseWeight = int(ability.Weight)
		}
		result = append(result, buildSnapshotDTO(ch, priority, baseWeight))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    result,
	})
}

// GetChannelHealthSnapshot 单渠道快照（管理员）
func GetChannelHealthSnapshot(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的渠道 ID"})
		return
	}
	ch, err := model.CacheGetChannel(id)
	if err != nil || ch == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "渠道不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    buildSnapshotDTO(ch, ch.GetPriority(), ch.GetWeight()),
	})
}

// ResetChannelHealth 重置某渠道的健康统计（管理员）
func ResetChannelHealth(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的渠道 ID"})
		return
	}
	channel_health.Reset(id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// AdjustChannelWeight 按固定步长调整渠道原始权重（管理员）
func AdjustChannelWeight(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的渠道 ID"})
		return
	}
	req := channelWeightAdjustRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	action := strings.TrimSpace(req.Action)
	if action != "raise" && action != "lower" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "不支持的操作"})
		return
	}
	ch, err := model.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	current := ch.GetWeight()
	next := current
	if action == "raise" {
		next += channelWeightAdjustStep
	} else {
		next -= channelWeightAdjustStep
		if next < 0 {
			next = 0
		}
	}
	weight := uint(next)
	ch.Weight = &weight
	if err := ch.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.InitChannelCache()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"channel_id":  ch.Id,
			"base_weight": ch.GetWeight(),
		},
	})
}

// ResetGroupChannelHealth 清空指定分组内所有渠道的健康统计（管理员）
func ResetGroupChannelHealth(c *gin.Context) {
	req := resetGroupChannelHealthRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	group := strings.TrimSpace(req.Group)
	if group == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "分组不能为空"})
		return
	}
	channelIDs := snapshotChannelIDs(group)
	for _, id := range channelIDs {
		channel_health.Reset(id)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"group":         group,
			"channel_count": len(channelIDs),
		},
	})
}

// UpdateGroupChannelAbility 更新分组维度下单渠道的 priority/weight（管理员）
func UpdateGroupChannelAbility(c *gin.Context) {
	req := updateGroupAbilityRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if req.Priority == nil && req.Weight == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "缺少更新字段"})
		return
	}
	if err := model.UpdateGroupChannelAbility(req.Group, req.ChannelId, req.Priority, req.Weight); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	model.InitChannelCache()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}
