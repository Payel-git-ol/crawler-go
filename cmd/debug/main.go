package main

import (
	"Fyne-on/pkg/database"
	"fmt"
	"log"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	prefixes := []string{"issue:", "pr:", "repo:", "contact:"}

	for _, prefix := range prefixes {
		count := 0
		err := db.IterateWithPrefix(prefix, func(key string, value []byte) error {
			count++
			if count <= 3 {
				fmt.Printf("Prefix '%s' - Sample key: %s (len=%d)\n", prefix, key, len(value))
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Prefix '%s' - Error: %v\n", prefix, err)
		}
		fmt.Printf("Prefix '%s' - Total count: %d\n\n", prefix, count)
	}
}
