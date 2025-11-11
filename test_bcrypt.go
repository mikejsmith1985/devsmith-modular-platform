package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	plainKey := "dsk_test_RK3jP9mL2nQ8vF7dW5tX"
	storedHash := "$2a$10$xVzK3qF7Qw9mL2nQ8vF7dOqW5tXhJfP9kR3jM6nL2nQ8vF7dW5tXh"
	
	fmt.Println("Testing bcrypt validation:")
	fmt.Printf("Plain key: %s\n", plainKey)
	fmt.Printf("Stored hash: %s\n", storedHash)
	fmt.Println()
	
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainKey))
	if err != nil {
		fmt.Println("❌ HASH MISMATCH:", err)
		fmt.Println()
		
		// Generate correct hash
		fmt.Println("Generating correct hash...")
		correctHash, err := bcrypt.GenerateFromPassword([]byte(plainKey), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println("Error generating hash:", err)
			return
		}
		
		fmt.Printf("✅ Correct hash for '%s':\n", plainKey)
		fmt.Println(string(correctHash))
		fmt.Println()
		fmt.Println("Update test_create_project.sql with this hash and re-run")
	} else {
		fmt.Println("✅ Hash matches! Validation should work.")
	}
}
