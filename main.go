package main

import (
	"bufio"
	"fmt"
	"google.golang.org/api/calendar/v3"
	"log"
	"os"
	"strings"
	"time"
)

import (
	"flag"
)

func main() {
	configPath := flag.Bool("config", false, "Flag to run the config, rather than sync command")
	flag.Parse()

	config, err := ReadFromConfig()

	if config != nil && (*configPath == true || err == nil) {

		if config.ApiKey == "" {
			ReadWriteConfig("XMatters API Key", "API_KEY")
		}

		if config.ApiSecret == "" {
			ReadWriteConfig("XMatters API Secret", "API_SECRET")
		}

		if config.Username == "" {
			ReadWriteConfig("XMatters Username", "USERNAME")
		}

		if *configPath == true {
			return
		}
	}

	calendarSrv, calendarId := PrepareCalendar()

	AddEventsToCalendar(calendarSrv, calendarId)
}

func ReadWriteConfig(label string, key string) {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(fmt.Sprintf("Enter your %s (leave blank to ignore): ", label))
	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)
	if value != "" {
		err := WriteToConfig(key, "\""+value+"\"")
		if err != nil {
			log.Fatalf("Error writing config: %v", err)
		}
	}
}

func PrepareCalendar() (*calendar.Service, string) {
	calendarName := "On Call"

	fmt.Printf("Looking for calendar '%s'\n", calendarName)
	calendarId, err := GetCalendarId(calendarName)
	if err != nil {
		log.Fatalf("Error looking for calendar: %v", err)
	}

	if calendarId == "" {
		fmt.Printf("Calendar '%s' not found, creating now\n", calendarName)
		calendarId, err = CreateCalendar(calendarName)

		if err != nil {
			log.Fatalf("Error creating calendar: %v", err)
		}
	}

	fmt.Printf("Calendar ID: %s\n", calendarId)

	calendarSrv, err := GetCalendarService()
	if err != nil {
		log.Fatalf("%v", err)
	}

	events, err := calendarSrv.Events.List(calendarId).TimeMin(time.Now().Format(time.RFC3339)).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}

	if len(events.Items) > 0 {
		fmt.Printf("Deleting %d existing events\n", len(events.Items))
		// Remove all existing events
		for _, event := range events.Items {
			err := calendarSrv.Events.Delete(calendarId, event.Id).Do()
			if err != nil {
				log.Fatalf("Unable to delete event: %v", err)
			}
		}
	}

	return calendarSrv, calendarId
}

func AddEventsToCalendar(calendarSrv *calendar.Service, calendarId string) {

	config, err := ReadFromConfig()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	if config.Username == "" {
		log.Fatalf("Username not set in config. Please run ./xmatters --config to configure")
	}

	fmt.Println("Getting my schedule")
	schedule, err := GetMySchedule(config.Username)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total: %d\n", schedule.Total)

	for _, shift := range schedule.Data {
		name := fmt.Sprintf("On Call %s", ParseTime(shift.Start).Format("2 Jan 06"))

		if shift.Replacing != "" {
			name = fmt.Sprintf("%s (replacing %s)", name, shift.Replacing)
		}

		event := &calendar.Event{
			Summary:     name,
			Description: name,
			Start: &calendar.EventDateTime{
				DateTime: shift.Start,
				TimeZone: "UTC",
			},
			End: &calendar.EventDateTime{
				DateTime: shift.End,
				TimeZone: "UTC",
			},
		}

		event, err = calendarSrv.Events.Insert(calendarId, event).Do()
		if err != nil {
			log.Fatalf("Unable to create event: %v", err)
		}

		fmt.Printf("Event created: %s\n", event.Summary)
	}
}
