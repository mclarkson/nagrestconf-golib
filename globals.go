package nrc

import (
	"fmt"
)

const (
	SUCCESS = 0
	ERROR   = 1
)

type HttpError struct {
	details string
}

func (e HttpError) Error() string {
	return fmt.Sprintf("%s", e.details)
}
