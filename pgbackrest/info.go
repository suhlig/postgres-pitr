package pgbackrest

import "encoding/json"

type Info struct {
	Name   string
	Status struct {
		Code    int
		Message string
	}
}

func ParseInfo(stdout string) ([]Info, error) {
	infos := make([]Info, 0)
	err := json.Unmarshal([]byte(stdout), &infos)
	return infos, err
}
