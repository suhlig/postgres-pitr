package postgres_pitr

// Runner executes commands
type Runner interface {
	Run(command string, args ...interface{}) (string, string, error)
}
