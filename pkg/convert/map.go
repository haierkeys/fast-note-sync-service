package convert

import (
	"fmt"
)

func MapAnyToMapStr(p map[string]interface{}) map[string]string {
	mapString := make(map[string]string)

	for key, value := range p {
		strKey := fmt.Sprintf("%v", key)
		strValue := fmt.Sprintf("%v", value)

		mapString[strKey] = strValue
	}
	return mapString
}
