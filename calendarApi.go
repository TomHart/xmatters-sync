package main

import (
	"context"
	"encoding/base64"
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
)

func GetCalendarService() (*calendar.Service, error) {
	ctx := context.Background()

	base64Str := "ewogICAgImluc3RhbGxlZCI6IHsKICAgICAgICAiY2xpZW50X2lkIjogIjc3MjM0NDEyNzQyMS05OWtsc3FtY3Q4cXRra3ZhY3A5YThkMTd0NXJvanJmZC5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsCiAgICAgICAgInByb2plY3RfaWQiOiAieG1hdHRlcnMtc3luYyIsCiAgICAgICAgImF1dGhfdXJpIjogImh0dHBzOi8vYWNjb3VudHMuZ29vZ2xlLmNvbS9vL29hdXRoMi9hdXRoIiwKICAgICAgICAidG9rZW5fdXJpIjogImh0dHBzOi8vb2F1dGgyLmdvb2dsZWFwaXMuY29tL3Rva2VuIiwKICAgICAgICAiYXV0aF9wcm92aWRlcl94NTA5X2NlcnRfdXJsIjogImh0dHBzOi8vd3d3Lmdvb2dsZWFwaXMuY29tL29hdXRoMi92MS9jZXJ0cyIsCiAgICAgICAgImNsaWVudF9zZWNyZXQiOiAiR09DU1BYLVJiR3ZQWVhZWm13R3pONjVPOXVpc2ZpZDdlcnMiLAogICAgICAgICJyZWRpcmVjdF91cmlzIjogWwogICAgICAgICAgICAiaHR0cDovL2xvY2FsaG9zdCIKICAgICAgICBdCiAgICB9Cn0="
	b, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		log.Fatalf("Error decoding base64 string: %v", err)
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

		err = SetGoogleToken(string(tokenBytes))
		if err != nil {
			log.Fatalf("Error saving token: %v", err)
		}
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Create a channel to receive the auth code
	codeChan := make(chan string, 1)

	var server *http.Server

	// Create temporary server
	server = &http.Server{
		Addr: ":80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the code from the URL parameters
			code := r.URL.Query().Get("code")
			if code != "" {
				codeChan <- code
				// Show success page to user
				fmt.Fprintf(w, "<h1>Success!</h1>You can close this window now.")
				// Shutdown server in a goroutine
				go func() {
					server.Shutdown(context.Background())
				}()
			}
		}),
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string

	authCode = <-codeChan
	//if _, err := fmt.Scan(&authCode); err != nil {
	//	log.Fatalf("Unable to read authorization code: %v", err)
	//}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile() (*oauth2.Token, error) {
	token, err := GetGoogleToken()
	if err != nil {
		return nil, err
	}

	tok := &oauth2.Token{}
	err = json.Unmarshal([]byte(token), tok)
	return tok, err
}
