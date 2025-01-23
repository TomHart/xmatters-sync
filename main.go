package main

import (
	"flag"
	"fmt"
	"google.golang.org/api/calendar/v3"
	"log"
	"strconv"
	"time"
)

func main() {
	SetupConfig()

	calendarSrv, calendarId := PrepareCalendar()

	AddEventsToCalendar(calendarSrv, calendarId)
}

func SetupConfig() {
	forceConfigure := flag.Bool("config", false, "Flag to run the config, rather than sync command")
	flag.Parse()

	EnsureXMattersDomainSet(*forceConfigure)
	EnsureApiKeySet(*forceConfigure)
	EnsureApiSecretSet(*forceConfigure)
	EnsureUsernameSet(*forceConfigure)
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

	username, _ := GetUsername()

	fmt.Println("Getting my schedule")
	schedule, err := GetMySchedule(username)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total: %d\n", schedule.Total)

	for _, shift := range schedule.Data {
		dayNumber, err := strconv.Atoi(ParseTime(shift.Start).Format("2"))
		if err != nil {
			log.Fatalf("Error converting day number: %v", err)
		}

		ordinal := GetOrdinal(dayNumber)
		name := fmt.Sprintf("On Call - %s", ParseTime(shift.Start).Format(fmt.Sprintf("Monday 2%s Jan 2006", ordinal)))

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

func GetOrdinal(n int) string {
	if n >= 11 && n <= 13 {
		return "th"
	}

	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}
