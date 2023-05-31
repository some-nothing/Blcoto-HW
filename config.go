package main

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigStruct struct {
	EndpointURL   string
	DatabaseURL   string
	StartPosition uint64
}

func LoadConfig() ConfigStruct {
	var config ConfigStruct

	f, _ := os.Open("config.json")
	defer f.Close()

	decoder := json.NewDecoder(f)
	err := decoder.Decode(&config)

	if err != nil {
		log.Fatal(err)
	}

	return config
}
