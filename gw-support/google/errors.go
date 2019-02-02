package google

import (
	"fmt"
)

// SupportEventNotFound is returned when the expected support event can not be found on the calendar.
type SupportEventNotFound struct{}

func (err SupportEventNotFound) Error() string {
	return "Could not find support event in calendar. Do you have access to the company calendar?"
}

// UnknownReturnType is returned when there is an unknown return type in the asyncGetKeybaseSecret function.
type UnknownReturnType struct {
	Data interface{}
}

func (err UnknownReturnType) Error() string {
	return fmt.Sprintf("Received unknown return type: %v", err.Data)
}
