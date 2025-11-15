package utilitytools

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Real tool implementations

func add(args map[string]interface{}) (interface{}, error) {
	// Required parameter: a (float64)
	a, ok := args["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("required parameter 'a' not found or wrong type")
	}

	// Required parameter: b (float64)
	b, ok := args["b"].(float64)
	if !ok {
		return nil, fmt.Errorf("required parameter 'b' not found or wrong type")
	}

	result := a + b
	return map[string]interface{}{
		"result": result,
	}, nil
}

func getCurrentTime(args map[string]interface{}) (interface{}, error) {
	currentTime := time.Now().Format(time.RFC3339)
	return map[string]interface{}{
		"time": currentTime,
	}, nil
}

func generateUUID(args map[string]interface{}) (interface{}, error) {
	newUUID := uuid.New().String()
	return map[string]interface{}{
		"uuid": newUUID,
	}, nil
}

func concatenateStrings(args map[string]interface{}) (interface{}, error) {
	// Required parameter: strings ([]interface{})
	stringsRaw, ok := args["strings"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("required parameter 'strings' not found or wrong type")
	}

	// Convert []interface{} to []string
	var stringSlice []string
	for _, s := range stringsRaw {
		if str, ok := s.(string); ok {
			stringSlice = append(stringSlice, str)
		}
	}

	// Optional parameter: separator (string)
	separator, _ := args["separator"].(string)
	if separator == "" {
		separator = " "
	}

	result := strings.Join(stringSlice, separator)
	return map[string]interface{}{
		"result": result,
	}, nil
}

func reverseString(args map[string]interface{}) (interface{}, error) {
	// Required parameter: text (string)
	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'text' not found or wrong type")
	}

	// Reverse the string
	runes := []rune(text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	result := string(runes)
	return map[string]interface{}{
		"result": result,
	}, nil
}
