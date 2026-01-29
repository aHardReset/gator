package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var fileName = "/.gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getFileName() (string, error) {
	filePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Home path reported as '%s'. %w", filePath, err)
	}
	return filePath + fileName, nil

}

func write(c *Config) error {
	fileFullName, err := getFileName()
	if err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("can't marshal %w", err)
	}
	if os.WriteFile(fileFullName, data, os.FileMode(os.O_WRONLY)) != nil {
		return fmt.Errorf("can't save to file %w", err)
	}

	return nil
}

func Read() (Config, error) {
	var config Config
	fileFullName, err := getFileName()
	if err != nil {
		return config, err
	}

	file, err := os.ReadFile(fileFullName)
	if err != nil {
		return config, fmt.Errorf("os Readfile error %w", err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, fmt.Errorf("can't unmarshal %w", err)
	}
	return config, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	if err := write(c); err != nil {
		return fmt.Errorf("can't unmarshal %w", err)
	}

	return nil
}
