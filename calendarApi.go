package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
	"log"
	"net/http"
	"os"
)

func GetCalendarService() (*calendar.Service, error) {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, errors.Join(errors.New("error reading client secret file"), err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope, calendar.CalendarEventsScope, people.UserinfoEmailScope, people.UserinfoProfileScope)
	if err != nil {
		return nil, errors.Join(errors.New("unable to parse client secret file to config"), err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, errors.Join(errors.New("unable to retrieve Calendar client"), err)
	}

	return srv, nil
}

func GetCalendarId(calendarName string) (string, error) {

	srv, err := GetCalendarService()
	if err != nil {
		return "", err
	}

	calendars, err := srv.CalendarList.List().Do()
	if err != nil {
		return "", err
	}

	for _, cld := range calendars.Items {
		if cld.Summary == calendarName {
			return cld.Id, nil
		}
	}

	return "", nil
}

func CreateCalendar(calendarName string) (string, error) {

	srv, err := GetCalendarService()
	if err != nil {
		return "", err
	}

	call := srv.Calendars.Insert(&calendar.Calendar{
		Summary: calendarName,
	})

	cld, err := call.Do()
	if err != nil {
		return "", err
	}

	return cld.Id, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tok, err := tokenFromFile()
	if err != nil {
		tok = getTokenFromWeb(config)
		tokenBytes, err := json.Marshal(tok)
		if err != nil {
			log.Fatalf("Error marshaling token: %v", err)
		}

		err = WriteToConfig("TOKEN", string(tokenBytes))
		if err != nil {
			log.Fatalf("Error saving token: %v", err)
		}
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile() (*oauth2.Token, error) {
	tokenValue, err := ReadFromConfig("TOKEN")
	if err != nil {
		return nil, err
	}

	tok := &oauth2.Token{}
	err = json.Unmarshal([]byte(tokenValue), tok)
	return tok, err
}
