package main

import (
"crypto/rand"
"encoding/base64"
"fmt"
"log"
"golang.org/x/crypto/bcrypt"
)

func main() {
randomBytes := make([]byte, 32)
_, err := rand.Read(randomBytes)
if err != nil {
erate random bytes: %v", err)
}
encoded := base64.RawURLEncoding.EncodeToString(randomBytes)
plainKey := "dsk_" + encoded
hashBytes, err := bcrypt.GenerateFromPassword([]byte(plainKey), bcrypt.DefaultCost)
if err != nil {
erate bcrypt hash: %v", err)
}
hash := string(hashBytes)
fmt.Println("=== API Key Generation ===")
fmt.Printf("Plain API Key: %s\n", plainKey)
fmt.Printf("Bcrypt Hash:   %s\n", hash)
fmt.Println()
fmt.Println("SQL UPDATE:")
fmt.Printf("UPDATE logs.projects SET api_key_hash = '%s' WHERE slug = 'load-test';\n", hash)
fmt.Println()
fmt.Println("Environment Variable:")
fmt.Printf("export LOGS_API_KEY='%s'\n", plainKey)
}
