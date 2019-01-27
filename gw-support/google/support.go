package google

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/gruntwork-io/prototypes/gw-support/logging"
)

const GruntworkCompanyCalendarID = "gruntwork.io_d12j5ekvrk7iu3a2ii7v6o753k@group.calendar.google.com"

func SupportNow(client *calendar.Service) error {
	logger := logging.GetProjectLogger()

	tNow := time.Now().Format(time.RFC3339)
	events, err := client.Events.List(GruntworkCompanyCalendarID).Q("Support:").
		ShowDeleted(false).SingleEvents(true).TimeMin(tNow).MaxResults(5).OrderBy("startTime").Do()
	if err != nil {
		return err
	}
	if len(events.Items) == 0 {
		return errors.WithStackTrace(SupportEventNotFound{})
	}
	supportNowEvent := events.Items[0]
	logger.Infof("Found calendar event \"%s\"", supportNowEvent.Summary)

	// ASSUMPTION: There should only be one attendee, which is the person on support
	currentSupport := supportNowEvent.Attendees[0]
	fmt.Printf("\n%s is on support now, until %s\n", currentSupport.Email, supportNowEvent.End.Date)
	fmt.Println("Upcoming rotation after that:")
	for _, event := range events.Items[1:] {
		fmt.Printf("\t%s to %s (%s)\n", event.Start.Date, event.End.Date, event.Summary)
	}
	return nil
}

func SupportNext(client *calendar.Service) error {
	logger := logging.GetProjectLogger()

	tNow := time.Now().Format(time.RFC3339)
	// Since the support events invite each person, the "Support:" event in YOUR calendar should be when you are on
	// support next
	events, err := client.Events.List("primary").Q("Support:").
		ShowDeleted(false).SingleEvents(true).TimeMin(tNow).MaxResults(5).OrderBy("startTime").Do()
	if err != nil {
		return err
	}
	if len(events.Items) == 0 {
		return errors.WithStackTrace(SupportEventNotFound{})
	}
	supportNextEvent := events.Items[0]
	logger.Infof("Found calendar event \"%s\"", supportNextEvent.Summary)

	fmt.Printf("\nYou are on support next %s until %s\n", supportNextEvent.Start.Date, supportNextEvent.End.Date)
	fmt.Println("Upcoming after that:")
	for _, event := range events.Items[1:] {
		fmt.Printf("\t%s to %s (%s)\n", event.Start.Date, event.End.Date, event.Summary)
	}
	return nil
}
