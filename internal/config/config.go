package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const ConfigFileName = "/.gatorconfig.json"

type Config struct {
	Db_url string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func ReadConfig() (*Config, error) {
	filepath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func getConfigFilePath() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homePath + ConfigFileName, nil
}

func (config *Config) SetUser(name string) error {
	if config == nil {
		config = &Config{}
	}
	if config.Current_user_name == name {
		fmt.Println("User is already set to:", name)
		return nil
	}
	config.Current_user_name = name

	filepath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(config)
	if err != nil {
		return err
	}
	return nil
}



