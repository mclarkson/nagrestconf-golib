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

type NrcQuery interface {
	RequiredOptions() []string
	Options() []string
	OptionsJson() string
	Show(bool, string)
	ShowJson(bool, bool, string)
	Get(string, string, string, string) error
	Post(string, string, string, string) error
}
