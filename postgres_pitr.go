package postgres_pitr

import "fmt"

// Runner executes commands
type Runner interface {
	Run(command string, args ...interface{}) (string, string, error)
}

// Error encapsulates information about failing run
type Error struct {
	Message        string
	Stdout, Stderr string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error: %s\nstderr: %s\nstdout: %s\n", e.Message, e.Stdout, e.Stderr)
}
