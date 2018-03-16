package main

import (
	"encoding/json"
	"os"
)

// Application config
var Config *config

type config struct {
	Port     string `json:"port"`
	Database struct {
		Host         string `json:"host"`
		Port         string `json:"port"`
		Username     string `json:"username"`
		Password     string `json:"password"`
		DatabaseName string `json:"database_name"`
		Ssl          string `json:"ssl"`
	} `json:"database"`
	Checklastblocktimeout string `json:"checklastblocktimeout"`
}

func init() {
	Config = loadConfig()
}

// LoadConfig load the config.json file and return its data
func loadConfig() *config {
	appConfig := &config{}

	configFile, err := os.Open("config.json")
	defer configFile.Close()
	if err != nil {
		panic(err)
	}

	err = json.NewDecoder(configFile).Decode(&appConfig)
	if err != nil {
		panic(err)
	}

	return appConfig
}
