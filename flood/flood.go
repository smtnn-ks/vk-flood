package flood

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type FloodControl_t struct {
	Rdb    *redis.Client
	Config FCConfig
}

type FCConfig struct {
	MaxRequest int64
	BanWindow  time.Duration
}

func (fc FloodControl_t) Check(ctx context.Context, userID int64) (bool, error) {
	var val *redis.IntCmd
	var expire *redis.BoolCmd
	uID := fmt.Sprint(userID)

	// Вместо ошибки `connection pool timeout` мы получаем val.Val() == 0
	for val == nil || val.Val() == 0 {
		fc.Rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			val = pipe.IncrBy(ctx, uID, 1)
			expire = pipe.Expire(ctx, uID, fc.Config.BanWindow)
			return nil
		})
		if err := val.Err(); err != nil {
			return false, err
		}
		if err := expire.Err(); err != nil {
			return false, err
		}
	}
	if val.Val() > fc.Config.MaxRequest {
		return false, nil
	}
	return true, nil
}

type FloodControl interface {
	Check(ctx context.Context, userID int64) (bool, error)
}
