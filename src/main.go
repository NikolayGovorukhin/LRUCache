package main

import (
	lru_cache "cache/src/lru_cache"
	"fmt"
	"time"
)

func main() {
	// Демонстрация работы
	cache := lru_cache.NewCacheTTL(3, -1)

	cache.Put(1, "str1")
	cache.Put(2, "str2")
	cache.Put(3, "str3")
	cache.Print()
	fmt.Println(cache.Get(3))
	fmt.Println(cache.Get(2))
	fmt.Println(cache.Get(1))
	fmt.Println(cache.Get(3))
	cache.Put(4, "str4")
	cache.Print()

	timer := time.NewTimer(time.Second * 5)
	<-timer.C

	cache.Print()
}
