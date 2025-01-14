package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetAPIKey() (string, error) {
	return ReadFromConfig("API_KEY")
}

func GetAPISecret() (string, error) {
	return ReadFromConfig("API_SECRET")
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

func ReadFromConfig(config string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}

	configFilePath := homeDir + "/.gdp/xmatters.conf"
	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(configFilePath), 0755); err != nil {
				return "", errors.Join(errors.New("error creating config directory"), err)
			}
			file, err = os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				return "", errors.Join(errors.New("error creating config file"), err)
			}
		} else {
			return "", errors.Join(errors.New("error opening config file"), err)
		}
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("error closing file: %v", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, config+"=") {
			return strings.TrimPrefix(line, config+"="), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", errors.Join(err, errors.New("error reading config file"))
	}

	return "", fmt.Errorf("%s not found in config file", config)
}
