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
	cache.Store("name", "Aboubakr Kouhadi", time.Second*5)
	cache.Store("email", "bryan@bryan.com", time.Second*5)
	cache.Store("age", 33, time.Second*5)
	cache.Store("profession", "English Teacher", time.Second*5)
	cache.Store("hobbies", "playing Guitar and soccer, swimming, and coding", time.Second*5)

	// fetch all data
	cache.Iterate(func(key string, value []byte) {
		fmt.Printf("key %s - Value %s", key, string(value))
	})

	// fetch a single piece of data
	if value, exists, err := cache.Fetch("name"); exists {
		if err == nil {
			fmt.Println("Fetched name: ", value)
		} else {
			fmt.Println("Error fetching name: ", err)
		}
	} else {
		fmt.Println("age does not exist or has expired or deleted...")
	}

	// Update a single piece of data
	err := cache.Update("name", "bryan bryan", time.Minute)
	if err != nil {
		fmt.Println("Update error:", err)
	}

	// fetch the updated value name
	if value, exists, err := cache.Fetch("name"); exists {
		if err == nil {
			fmt.Println("Fetched updated name:", value)
		} else {
			fmt.Println("Error fetching name:", err)
		}
	} else {
		fmt.Println("name does not exist or has expired or deleted...")
	}

	// Delete the value
	cache.Delete("profession")

	// fetch the deleted value profession
	if value, exists, err := cache.Fetch("profession"); exists {
		if err == nil {
			fmt.Println("Fetched profession after it's been deleted:", value)
		} else {
			fmt.Println("Error fetching profession:", err)
		}
	} else {
		fmt.Println("profession does not exist or has expired or deleted...")
	}
	// clean up all data
	cache.CleanupAll()
	fmt.Println("data has been cleaned up....")
	// fetch age after clean up all data
	value, exists, err := cache.Fetch("age")
	if err != nil {
		fmt.Println("Fetch error:", err)
	}
	fmt.Println(value, exists)

	// fetch all data after cleanup
	cache.Iterate(func(key string, value []byte) {
		fmt.Printf("key: %s -  Value: %s", key, string(value))
	})
	
	// store again some data
	cache.Store("test", "automatic deletion after expiration", time.Second)
	// wait for some time and
	time.Sleep(time.Millisecond * 1200)
	// fetch expired data
	if value, exists, err := cache.Fetch("test"); exists {
		if err == nil {
			fmt.Println("Fetched expired test:", value)
		} else {
			fmt.Println("Error fetching test:", err)
		}
	} else {
		fmt.Println("test does not exist or has expired or deleted...")
	}
	fmt.Println("END")
}
