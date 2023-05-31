package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	EndpointURL string
	DatabaseURL string
}

func LoadConfig() Config {
	var config Config

	f, _ := os.Open("config.json")
	defer f.Close()

	decoder := json.NewDecoder(f)
	err := decoder.Decode(&config)

	if err != nil {
		log.Fatal(err)
	}

	return config
}
