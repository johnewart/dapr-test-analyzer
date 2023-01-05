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

	if db := ingestion.NewRecordDB(); db == nil {
		fmt.Println("Unable to open DB...")
	} else {
		if err := ingestion.ExtractResults(db); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
}
