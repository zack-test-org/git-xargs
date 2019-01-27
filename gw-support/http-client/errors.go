package http_client

import (
	"fmt"
)

type MaxRetriesError struct {
	Action string
}

func (err MaxRetriesError) Error() string {
	return fmt.Sprintf("Exceeded maximum number of retries performing %s", err.Action)
}
