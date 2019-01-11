package walg

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ParseListOutput parses the output of `wal-g backup-list` into an `Info` struct
func ParseListOutput(output string) (*Info, error) {
	lines := strings.Split(output, "\n")

	switch len(lines) {
	case 0:
		return nil, errors.New("Error listing backups")
	case 1:
		return nil, errors.New("Parse error: not enough lines")
	default:
		path, err := parsePathLine(lines[0])

		if err != nil {
			return nil, err
		}

		backups, err := parseBackupLines(lines[2:])

		if err != nil {
			return nil, err
		}

		return &Info{
			Path:    path,
			Backups: *backups,
		}, nil
	}
}

func parsePathLine(line string) (string, error) {
	parts := strings.Split(line, ":")

	if 2 != len(parts) {
		return "", fmt.Errorf("Error parsing path line '%s'; expected 2 parts, but found %d", line, len(parts))
	}

	return strings.TrimSpace(parts[1]), nil
}

func parseBackupLines(lines []string) (*[]Backup, error) {
	backups := make([]Backup, len(lines))

	for i, line := range lines {
		if 0 == len(line) {
			backups = backups[:len(backups)-1]
			continue
		}

		backup, err := parseBackupLine(line)

		if err != nil {
			return nil, err
		}

		backups[i] = *backup
	}

	return &backups, nil
}

func parseBackupLine(line string) (*Backup, error) {
	fields := strings.Split(line, " ")

	if 3 != len(fields) {
		return nil, fmt.Errorf("Error parsing backup line '%s'; expected 3 parts, but found %d", line, len(fields))
	}

	lastModified, err := time.Parse(time.RFC3339, fields[1])

	if err != nil {
		return nil, err
	}

	return &Backup{
		Name:                  fields[0],
		LastModified:          lastModified,
		WalSegmentBackupStart: fields[2],
	}, nil
}
