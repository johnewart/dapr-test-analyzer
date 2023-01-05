package main

import (
	"github.com/joho/godotenv"
	"log"
	"test-analyzer/web"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	web.ServeHTTP()
}
