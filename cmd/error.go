package cmd

import "fmt"

type CLIError struct {
	ctxErr error
}

func NewCLIError(err error) *CLIError {
	return &CLIError{ctxErr: err}
}

func (err CLIError) Error() string {
	return fmt.Sprintf("error on reading CLI flags: %s", err.ctxErr.Error())
}
