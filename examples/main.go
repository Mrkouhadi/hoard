package main

import (
	"fmt"
	"time"

	"github.com/mrkouhadi/hoard"
)

func main() {
	// create a cache with 5 shards, maximum of 10000 items per shard, and a cleanup interval of 10 seconds
	cache := hoard.NewCache(5, 10000, time.Second*10)

	// store items
	cache.Store("age", 33, time.Second*3)
	cache.Store("name", "aboubakr", time.Second*5)

	// add a small delay to ensure the items are stored
	time.Sleep(100 * time.Millisecond)

	// fetch and print the value of "age" (should exist)
	if value, exists, err := cache.Fetch("age"); exists {
		if err == nil {
			fmt.Println("Fetched age:", value)
		} else {
			fmt.Println("Error fetching age:", err)
		}
	} else {
		fmt.Println("age does not exist or has expired")
	}

	// wait for 4 seconds to let "age" expire but "name" still exists
	time.Sleep(4 * time.Second)

	// Fetch and print the value of "age" again (should be expired)
	if value, exists, err := cache.Fetch("age"); exists {
		if err == nil {
			fmt.Println("Fetched age:", value)
		} else {
			fmt.Println("Error fetching age:", err)
		}
	} else {
		fmt.Println("age does not exist or has expired")
	}

	// fetch and print the value of "name" (should still exist)
	if value, exists, err := cache.Fetch("name"); exists {
		if err == nil {
			fmt.Println("Fetched name:", value)
		} else {
			fmt.Println("Error fetching name:", err)
		}
	} else {
		fmt.Println("name does not exist or has expired")
	}

	// wait for 2 more seconds to let "name" expire
	time.Sleep(2 * time.Second)

	// fetch and print the value of "name" again (should be expired)
	if value, exists, err := cache.Fetch("name"); exists {
		if err == nil {
			fmt.Println("Fetched name:", value)
		} else {
			fmt.Println("Error fetching name:", err)
		}
	} else {
		fmt.Println("name does not exist or has expired")
	}
}
