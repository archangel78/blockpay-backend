package config

import (
	"os"
	"io/ioutil"
	"encoding/json"
)

type DbConfig struct {
	Protocol     string `json:"protocol"`
	Hostname     string `json:"hostname"`
	Port         string `json:"port"`
	Username     string
	Password     string
	DatabaseName string `json:"databaseName"`
}

func GetConfig(configPath string) (*DbConfig, error) {
	configFile, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	dbConfig := DbConfig{}
	err = json.Unmarshal(configFile, &dbConfig)

	if err != nil {
		return nil, err
	}

	dbConfig.Username = os.Getenv("MYSQL_USERNAME")
	dbConfig.Password = os.Getenv("MYSQL_PASSWORD")

	return &dbConfig, nil
}
