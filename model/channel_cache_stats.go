package model

import (
	"math"
	"time"

	"gorm.io/gorm/clause"
)

type ChannelCacheStat struct {
	Id                   int       `json:"id" gorm:"primaryKey;autoIncrement"`
	ChannelId            int       `json:"channel_id" gorm:"uniqueIndex:idx_channel_date;not null"`
	StatDate             string    `json:"stat_date" gorm:"type:varchar(10);uniqueIndex:idx_channel_date;not null"`
	TotalRequests        int64     `json:"total_requests" gorm:"default:0"`
	CacheHitRequests     int64     `json:"cache_hit_requests" gorm:"default:0"`
	TotalPromptTokens    int64     `json:"total_prompt_tokens" gorm:"default:0"`
	CacheHitTokens       int64     `json:"cache_hit_tokens" gorm:"default:0"`
	CacheCreationTokens  int64     `json:"cache_creation_tokens" gorm:"default:0"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ChannelCacheStatResult struct {
	ChannelId           int     `json:"channel_id"`
	ChannelName         string  `json:"channel_name"`
	TotalRequests       int64   `json:"total_requests"`
	CacheHitRequests    int64   `json:"cache_hit_requests"`
	HitRate             float64 `json:"hit_rate"`
	TotalPromptTokens   int64   `json:"total_prompt_tokens"`
	CacheHitTokens      int64   `json:"cache_hit_tokens"`
	TokenHitRatio       float64 `json:"token_hit_ratio"`
	CacheCreationTokens int64   `json:"cache_creation_tokens"`
}

func UpsertChannelCacheStats(stats []*ChannelCacheStat) error {
	if len(stats) == 0 {
		return nil
	}
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "channel_id"}, {Name: "stat_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_requests", "cache_hit_requests", "total_prompt_tokens", "cache_hit_tokens", "cache_creation_tokens", "updated_at"}),
	}).Create(&stats).Error
}

func GetChannelCacheStats(days int) ([]*ChannelCacheStatResult, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	var rows []struct {
		ChannelId           int
		ChannelName         string
		TotalRequests       int64
		CacheHitRequests    int64
		TotalPromptTokens   int64
		CacheHitTokens      int64
		CacheCreationTokens int64
	}

	err := DB.Table("channel_cache_stats").
		Select("channel_cache_stats.channel_id, channels.name as channel_name, "+
			"SUM(channel_cache_stats.total_requests) as total_requests, "+
			"SUM(channel_cache_stats.cache_hit_requests) as cache_hit_requests, "+
			"SUM(channel_cache_stats.total_prompt_tokens) as total_prompt_tokens, "+
			"SUM(channel_cache_stats.cache_hit_tokens) as cache_hit_tokens, "+
			"SUM(channel_cache_stats.cache_creation_tokens) as cache_creation_tokens").
		Joins("LEFT JOIN channels ON channels.id = channel_cache_stats.channel_id").
		Where("channel_cache_stats.stat_date >= ?", cutoff).
		Group("channel_cache_stats.channel_id, channels.name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	results := make([]*ChannelCacheStatResult, 0, len(rows))
	for _, r := range rows {
		res := &ChannelCacheStatResult{
			ChannelId:           r.ChannelId,
			ChannelName:         r.ChannelName,
			TotalRequests:       r.TotalRequests,
			CacheHitRequests:    r.CacheHitRequests,
			TotalPromptTokens:   r.TotalPromptTokens,
			CacheHitTokens:      r.CacheHitTokens,
			CacheCreationTokens: r.CacheCreationTokens,
		}
		if r.TotalRequests > 0 {
			res.HitRate = math.Round(float64(r.CacheHitRequests)/float64(r.TotalRequests)*1000) / 1000
		}
		if r.TotalPromptTokens > 0 {
			res.TokenHitRatio = math.Round(float64(r.CacheHitTokens)/float64(r.TotalPromptTokens)*1000) / 1000
		}
		results = append(results, res)
	}
	return results, nil
}
