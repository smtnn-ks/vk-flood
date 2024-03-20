package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"vk-flood/flood"

	"github.com/redis/go-redis/v9"
)

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func main() {
	fc := flood.FloodControl_t{
		Config: flood.FCConfig{
			MaxRequest: 10,
			BanWindow:  time.Second * 10,
		},
	}
	fc.Rdb = initRedis()
	if fc.Rdb != nil {
		var wg sync.WaitGroup
		iters := 100000
		wg.Add(iters)
		var doneOps int64 = 0
		for i := 0; i < iters; i++ {
			fmt.Printf("\r%d", i)
			go func(fl *flood.FloodControl_t) {
				ok, err := fl.Check(context.Background(), 1)
				if err != nil {
					fmt.Println("ERROR ", err)
				}
				if ok {
					atomic.AddInt64(&doneOps, 1)
				}
				wg.Done()
			}(&fc)
		}
		wg.Wait()
		fmt.Printf("\nOpts done: %d\n", doneOps)
	} else {
		panic("fl.Rdb is nil")
	}
}
