package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()

	// 아이템 갯수 170개 생성
	key := "item-2"
	client.Set(ctx, key, "183", -1)

	// lua script
	var itemCount = redis.NewScript(`
	local key = KEYS[1]
	local cur = tonumber(redis.call("GET", key) or "-1")
	
	if cur > 0 then
	  return redis.call("INCRBY", key, -1)
	else
	  return -1
	end
	`)

	// 40의 컨슈머가 동시요청 상황 가정
	ch := make(chan bool, 40)
	for i := 0; i < 200; i++ {
		go func(i int, ch chan bool) {
			num, err := itemCount.Run(ctx, client, []string{key}).Int64()
			if err != nil {
				panic(err)
			}

			if num < 0 {
				fmt.Println("failed ", num, i)
				ch <- false
			} else {
				fmt.Println("num is ", num)
				ch <- true
			}
		}(i, ch)
	}

	var failedCount int
	for i := 0; i < 200; i++ {
		success := <-ch
		if !success {
			failedCount++
		}
	}

	fmt.Printf("success %d, failed %d\n", 200-failedCount, failedCount)
}
