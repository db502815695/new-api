package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	cacheStatsBatchSize = 1000
)

var (
	cacheStatsOnce    sync.Once
	cacheStatsRunning atomic.Bool
)

func StartChannelCacheStatsTask() {
	cacheStatsOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		gopool.Go(func() {
			logger.LogInfo(context.Background(), "channel cache stats task started")
			for {
				now := time.Now()
				// next 07:00 local time
				next := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, now.Location())
				if !next.After(now) {
					next = next.Add(24 * time.Hour)
				}
				timer := time.NewTimer(next.Sub(now))
				<-timer.C
				timer.Stop()
				runChannelCacheStatsOnce()
			}
		})
	})
}

func runChannelCacheStatsOnce() {
	if !cacheStatsRunning.CompareAndSwap(false, true) {
		return
	}
	defer cacheStatsRunning.Store(false)

	ctx := context.Background()
	yesterday := time.Now().AddDate(0, 0, -1)
	statDate := yesterday.Format("2006-01-02")
	startTime := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location()).Unix()
	endTime := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999999999, yesterday.Location()).Unix()

	type accumulator struct {
		totalRequests       int64
		cacheHitRequests    int64
		totalPromptTokens   int64
		cacheHitTokens      int64
		cacheCreationTokens int64
	}
	acc := make(map[int]*accumulator)

	offset := 0
	for {
		var logs []model.Log
		err := model.DB.
			Select("channel_id, prompt_tokens, other").
			Where("created_at >= ? AND created_at <= ? AND type = ?", startTime, endTime, model.LogTypeConsume).
			Limit(cacheStatsBatchSize).
			Offset(offset).
			Find(&logs).Error
		if err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("channel cache stats: query logs failed: %v", err))
			return
		}
		if len(logs) == 0 {
			break
		}

		for _, log := range logs {
			a, ok := acc[log.ChannelId]
			if !ok {
				a = &accumulator{}
				acc[log.ChannelId] = a
			}
			a.totalRequests++
			a.totalPromptTokens += int64(log.PromptTokens)

			if log.Other != "" {
				var otherMap map[string]interface{}
				if err := common.UnmarshalJsonStr(log.Other, &otherMap); err == nil {
					if v, ok := otherMap["cache_tokens"]; ok {
						cacheTokens := toInt64(v)
						if cacheTokens > 0 {
							a.cacheHitRequests++
							a.cacheHitTokens += cacheTokens
						}
					}
					if v, ok := otherMap["cache_creation_tokens"]; ok {
						a.cacheCreationTokens += toInt64(v)
					}
				}
			}
		}

		offset += len(logs)
		if len(logs) < cacheStatsBatchSize {
			break
		}
	}

	if len(acc) == 0 {
		logger.LogInfo(ctx, fmt.Sprintf("channel cache stats: no data for %s", statDate))
		return
	}

	stats := make([]*model.ChannelCacheStat, 0, len(acc))
	for channelId, a := range acc {
		stats = append(stats, &model.ChannelCacheStat{
			ChannelId:           channelId,
			StatDate:            statDate,
			TotalRequests:       a.totalRequests,
			CacheHitRequests:    a.cacheHitRequests,
			TotalPromptTokens:   a.totalPromptTokens,
			CacheHitTokens:      a.cacheHitTokens,
			CacheCreationTokens: a.cacheCreationTokens,
			UpdatedAt:           time.Now(),
		})
	}

	if err := model.UpsertChannelCacheStats(stats); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("channel cache stats: upsert failed: %v", err))
		return
	}
	logger.LogInfo(ctx, fmt.Sprintf("channel cache stats: aggregated %d channels for %s", len(stats), statDate))
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	}
	return 0
}
