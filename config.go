package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	ApiKey    string `toml:"API_KEY"`
	ApiSecret string `toml:"API_SECRET"`
	Username  string
	Token     string
}

func GetAPIKey() (string, error) {
	config, err := ReadFromConfig()
	if err != nil {
		return "", err
	}

	if config.ApiKey == "" {
		return "", errors.New("API_KEY not found in config file. Please ensure API_KEY, API_SECRET, and USERNAME are set in ~/.gdp/xmatters.conf")
	}

	return config.ApiKey, nil
}

func GetAPISecret() (string, error) {
	config, err := ReadFromConfig()
	if err != nil {
		return "", err
	}

	if config.ApiSecret == "" {
		return "", errors.New("API_SECRET not found in config file. Please ensure API_KEY, API_SECRET, and USERNAME are set in ~/.gdp/xmatters.conf")
	}

	return config.ApiSecret, nil
}

func WriteToConfig(key string, value string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %v", err)
	}

	configFilePath := homeDir + "/.gdp/xmatters.conf"
	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(configFilePath), 0755); err != nil {
				return errors.Join(errors.New("error creating config directory"), err)
			}
			_, err = os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				return errors.Join(errors.New("error creating config file"), err)
			}
		} else {
			return errors.Join(errors.New("error opening config file"), err)
		}
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key+"=") {
			lines = append(lines, key+"="+value)
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return errors.Join(err, errors.New("error reading config file"))
	}

	if !strings.Contains(strings.Join(lines, "\n"), key+"="+value) {
		lines = append(lines, key+"="+value)
	}

	file, err = os.OpenFile(configFilePath, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Join(errors.New("error opening config file for writing"), err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("error closing file: %v", err)
		}
	}(file)

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return errors.Join(errors.New("error writing to config file"), err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func ReadFromConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %v", err)
	}

	configFilePath := homeDir + "/.gdp/xmatters.conf"
	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(configFilePath), 0755); err != nil {
				return nil, errors.Join(errors.New("error creating config directory"), err)
			}
			file, err = os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				return nil, errors.Join(errors.New("error creating config file"), err)
			}
		} else {
			return nil, errors.Join(errors.New("error opening config file"), err)
		}
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("error closing file: %v", err)
		}
	}(file)

	tomlData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var conf Config
	_, err = toml.Decode(string(tomlData), &conf)
	if err != nil {
		return nil, fmt.Errorf("error decoding config file: %v", err)
	}

	return &conf, nil
}
