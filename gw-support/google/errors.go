package google

type SupportEventNotFound struct{}

func (err SupportEventNotFound) Error() string {
	return "Could not find support event in calendar. Do you have access to the company calendar?"
}
