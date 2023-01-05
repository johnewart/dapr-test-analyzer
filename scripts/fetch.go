package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"test-analyzer/ingestion"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := ingestion.FetchResults(900, 50); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}
