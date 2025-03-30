package utils

import (
	"log"
	"strconv"
)

func ConvertResultsToMapStringInterface(results map[int]int) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range results {
		result[strconv.Itoa(key)] = value
	}
	return result
}

func ConvertResultsToIntMap(data map[interface{}]interface{}) map[int]int {
	results := make(map[int]int)
	for key, value := range data {
		keyStr := key.(string)
		keyInt, _ := strconv.Atoi(keyStr)
		results[keyInt] = int(value.(int64))
	}
	return results
}

func ConvertToStringSlice(data []interface{}) []string {
	result := make([]string, len(data))
	for i, v := range data {
		result[i] = v.(string)
	}
	return result
}

func ConvertMapStringInterfaceToResults(data map[string]interface{}) map[int]int {
	results := make(map[int]int)
	if data == nil {
		return results
	}
	for key, value := range data {
		keyInt, err := strconv.Atoi(key)
		if err != nil {
			log.Printf("Error converting key '%s' to int: %v\n", key, err) //erroring
			continue
		}
		valueInt64, ok := value.(int64)
		if !ok {
			log.Printf("Error converting value for key '%s' to int64: value is of type %T\n", key, value) //erroring
			continue
		}
		valueInt := int(valueInt64)

		results[keyInt] = valueInt
	}

	return results
}
