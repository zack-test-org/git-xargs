package google

import (
	"fmt"
	"regexp"
	"time"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/gruntwork-io/prototypes/gw-support/logging"
)

const (
	// You can get the ID of any calendar by viewing the settings of any calendar.
	// For example, this calendar is the "Company Calendar". You can get the ID by adding this calendar to your calendar
	// in the "Add Calendar" box on the left side bar.
	// Then, click the kebab icon (the three vertical dots) and select settings.
	// Finally, scroll to "Integrate Calendar" and the ID should be under "Calendar ID"
	GruntworkCompanyCalendarID = "gruntwork.io_d12j5ekvrk7iu3a2ii7v6o753k@group.calendar.google.com"

	// This is a Google Calendar API documentation.
	UserCalendarID = "primary"

	// These are assumptions based on how we construct the support events. Support events are created in the
	// Gruntwork "Company Calendar", with the name "Support (Primary): PERSON" with only that person invited.
	PrimarySupportQuery = "Support (Primary):"
	BackupSupportQuery  = "Support (Backup):"
)

// getSupportEvents will return the list of support events (up to 5) that match the query, filtering from now and
// returning in chronological order (closest event first).
func getSupportEvents(eventsList *calendar.EventsListCall, query string) (*calendar.Events, error) {
	tNow := time.Now().Format(time.RFC3339)
	events, err := eventsList.Q(query).ShowDeleted(false).SingleEvents(true).TimeMin(tNow).MaxResults(5).OrderBy("startTime").Do()
	if err != nil {
		return events, errors.WithStackTrace(err)
	}
	if len(events.Items) == 0 {
		return events, errors.WithStackTrace(SupportEventNotFound{})
	}
	return events, nil
}

// interleavedPrimaryBackupEvents will query for both primary and backup events, and then return a combined list in
// chronological order.
func interleavedPrimaryBackupEvents(eventsList *calendar.EventsListCall) ([]*calendar.Event, error) {
	primaryEventsResp, err := getSupportEvents(eventsList, PrimarySupportQuery)
	if err != nil {
		return nil, err
	}
	primaryEvents := primaryEventsResp.Items
	backupEventsResp, err := getSupportEvents(eventsList, BackupSupportQuery)
	if err != nil {
		return nil, err
	}
	backupEvents := backupEventsResp.Items

	interleaved := []*calendar.Event{}

	var event *calendar.Event
	for len(primaryEvents) > 0 || len(backupEvents) > 0 {
		if len(primaryEvents) == 0 {
			// pop left from backupEvents because there are no more primary events
			event, backupEvents = backupEvents[0], backupEvents[1:]
		} else if len(backupEvents) == 0 {
			// pop left from primaryEvents because there are no more backup events
			event, primaryEvents = primaryEvents[0], primaryEvents[1:]
		} else if backupEvents[0].Start.Date < primaryEvents[0].Start.Date {
			// pop left from backupEvents because it is earlier
			event, backupEvents = backupEvents[0], backupEvents[1:]
		} else {
			// pop left from primaryEvents because it is earlier
			event, primaryEvents = primaryEvents[0], primaryEvents[1:]
		}
		interleaved = append(interleaved, event)
	}
	return interleaved, nil
}

// getSupportType uses the assumption that our support events are named "Support (TYPE):" to extract the type.
func getSupportType(summaryText string) string {
	re := regexp.MustCompile(`Support \((Primary|Backup)\):`)
	return re.FindStringSubmatch(summaryText)[1]
}

// SupportNow will print out who is on support now, and in the next 4 weeks.
func SupportNow(client *calendar.Service) error {
	logger := logging.GetProjectLogger()

	primaryEvents, err := getSupportEvents(client.Events.List(GruntworkCompanyCalendarID), PrimarySupportQuery)
	if err != nil {
		return err
	}
	primarySupportNowEvent := primaryEvents.Items[0]
	logger.Infof("Found calendar event \"%s\" for primary", primarySupportNowEvent.Summary)

	backupEvents, err := getSupportEvents(client.Events.List(GruntworkCompanyCalendarID), BackupSupportQuery)
	if err != nil {
		return err
	}
	backupSupportNowEvent := backupEvents.Items[0]
	logger.Infof("Found calendar event \"%s\" for backup", backupSupportNowEvent.Summary)

	// ASSUMPTION: There should only be one attendee, which is the person on support
	fmt.Println("\nUpcoming rotation:")
	currentPrimarySupport := primarySupportNowEvent.Attendees[0]
	currentBackupSupport := backupSupportNowEvent.Attendees[0]
	fmt.Printf(
		"\t%s is on primary support now (%s is backup), until %s\n",
		currentPrimarySupport.Email,
		currentBackupSupport.Email,
		primarySupportNowEvent.End.Date,
	)
	for idx, event := range primaryEvents.Items {
		backupEvent := backupEvents.Items[idx]
		primarySupport := event.Attendees[0]
		backupSupport := backupEvent.Attendees[0]
		fmt.Printf(
			"\t%s to %s (Primary: %s\tBackup: %s)\n",
			event.Start.Date,
			event.End.Date,
			primarySupport.Email,
			backupSupport.Email,
		)
	}
	return nil
}

// SupportNext prints out when the authenticated user is on support next, and the 4 support events after that.
func SupportNext(client *calendar.Service) error {
	// Since the support events invite each person, the "Support (Primary):" event in YOUR calendar should be when you are on
	// support next
	eventsList := client.Events.List(UserCalendarID)
	allEvents, err := interleavedPrimaryBackupEvents(eventsList)
	if err != nil {
		return errors.WithStackTrace(err)
	}

	fmt.Println("\nYour upcoming support:")
	for _, event := range allEvents {
		supportType := getSupportType(event.Summary)
		fmt.Printf("\t%s to %s (%s)\n", event.Start.Date, event.End.Date, supportType)
	}
	return nil
}
