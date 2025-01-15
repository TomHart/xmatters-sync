package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/keybase/go-keychain"
	"log"
	"os"
	"strings"
)

func EnsureApiKeySet(forceConfigure bool) {
	apiKey, err := GetApiKey()
	if err != nil {
		var valErr *NoValue
		if !errors.As(err, &valErr) {
			log.Fatalf("error getting api key: %v", valErr.Err)
		}

		ReadFromUser("XMatters API Key", "api_key")
		return
	}

	if apiKey == "" || forceConfigure == true {
		ReadFromUser("XMatters API Key", "api_key")
	}
}

func EnsureApiSecretSet(forceConfigure bool) {
	apiKey, err := GetApiSecret()
	if err != nil {
		var valErr *NoValue
		if !errors.As(err, &valErr) {
			log.Fatalf("error getting api secret: %v", valErr.Err)
		}

		ReadFromUser("XMatters API Secret", "api_secret")
		return
	}

	if apiKey == "" || forceConfigure == true {
		ReadFromUser("XMatters API Secret", "api_secret")
	}
}

func EnsureXMattersDomainSet(forceConfigure bool) {
	apiKey, err := GetXMattersDomain()
	if err != nil {
		var valErr *NoValue
		if !errors.As(err, &valErr) {
			log.Fatalf("error getting username: %v", valErr.Err)
		}

		ReadFromUser("XMatters Domain (abc.xmatters.sky for example)", "domain")
		return
	}

	if apiKey == "" || forceConfigure == true {
		ReadFromUser("XMatters Domain (abc.xmatters.sky for example)", "domain")
	}
}

func EnsureUsernameSet(forceConfigure bool) {
	apiKey, err := GetUsername()
	if err != nil {
		var valErr *NoValue
		if !errors.As(err, &valErr) {
			log.Fatalf("error getting username: %v", valErr.Err)
		}

		ReadFromUser("XMatters Username", "username")
		return
	}

	if apiKey == "" || forceConfigure == true {
		ReadFromUser("XMatters Username", "username")
	}
}

type NoValue struct {
	Err error
}

func (r *NoValue) Error() string {
	return fmt.Sprintf("no api key set: err %v", r.Err)
}

func GetApiKey() (string, error) {
	// Doesn't error if it's not set
	key, err := keychain.GetGenericPassword("XMatters Sync", "api_key", "", "xmatters")
	keyStr := string(key)

	if err != nil {
		return "", &NoValue{Err: err}
	}

	if keyStr == "" {
		return "", &NoValue{}
	}

	return keyStr, nil
}

func GetApiSecret() (string, error) {
	// Doesn't error if it's not set
	secret, err := keychain.GetGenericPassword("XMatters Sync", "api_secret", "", "xmatters")
	secretStr := string(secret)

	if err != nil {
		return "", &NoValue{Err: err}
	}

	if secretStr == "" {
		return "", &NoValue{}
	}

	return secretStr, nil
}

func GetXMattersDomain() (string, error) {
	// Doesn't error if it's not set
	domain, err := keychain.GetGenericPassword("XMatters Sync", "domain", "", "xmatters")
	domainStr := string(domain)

	if err != nil {
		return "", &NoValue{Err: err}
	}

	if domainStr == "" {
		return "", &NoValue{}
	}

	return domainStr, nil
}

func GetUsername() (string, error) {
	// Doesn't error if it's not set
	username, err := keychain.GetGenericPassword("XMatters Sync", "username", "", "xmatters")
	usernameStr := string(username)

	if err != nil {
		return "", &NoValue{Err: err}
	}

	if usernameStr == "" {
		return "", &NoValue{}
	}

	return usernameStr, nil
}

func GetGoogleToken() (string, error) {
	// Doesn't error if it's not set
	username, err := keychain.GetGenericPassword("XMatters Sync", "google_token", "", "xmatters")
	usernameStr := string(username)

	if err != nil {
		return "", &NoValue{Err: err}
	}

	if usernameStr == "" {
		return "", &NoValue{}
	}

	return usernameStr, nil
}

func SetGoogleToken(token string) error {
	_ = keychain.DeleteGenericPasswordItem("XMatters Sync", "google_token")

	item := keychain.NewGenericPassword("XMatters Sync", "google_token", "", []byte(token), "xmatters")
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	err := keychain.AddItem(item)

	return err
}

func ReadFromUser(label string, key string) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(fmt.Sprintf("Enter your %s: ", label))
		value, _ := reader.ReadString('\n')
		value = strings.TrimSpace(value)
		if value != "" {
			_ = keychain.DeleteGenericPasswordItem("XMatters Sync", key)

			item := keychain.NewGenericPassword("XMatters Sync", key, "", []byte(value), "xmatters")
			item.SetSynchronizable(keychain.SynchronizableNo)
			item.SetAccessible(keychain.AccessibleWhenUnlocked)
			err := keychain.AddItem(item)
			if errors.Is(err, keychain.ErrorDuplicateItem) {
				log.Fatalf("wasn't able to save the key: %v", err)
			}

			return
		}
	}
}
