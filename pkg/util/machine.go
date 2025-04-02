package util

import (
	"fmt"

	"github.com/denisbrodbeck/machineid"
)

func GetMachineID() string {
	id, err := machineid.ID()
	if err != nil {
		fmt.Println("Failed to get machine ID")
		return ""
	}
	return id
}
