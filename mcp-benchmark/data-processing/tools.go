package dataprocessing

import (
	"fmt"
	"sort"
)

func filterArray(args map[string]interface{}) (interface{}, error) {
	data, ok := args["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("data must be an array")
	}

	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	value, ok := args["value"].(float64)
	if !ok {
		return nil, fmt.Errorf("value must be a number")
	}

	filtered := []interface{}{}
	for _, item := range data {
		num, ok := item.(float64)
		if !ok {
			continue
		}

		include := false
		switch operation {
		case "gt":
			include = num > value
		case "lt":
			include = num < value
		case "eq":
			include = num == value
		case "gte":
			include = num >= value
		case "lte":
			include = num <= value
		}

		if include {
			filtered = append(filtered, num)
		}
	}

	return map[string]interface{}{
		"result": filtered,
		"count":  len(filtered),
	}, nil
}

func mapArray(args map[string]interface{}) (interface{}, error) {
	data, ok := args["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("data must be an array")
	}

	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	mapped := make([]interface{}, len(data))
	for i, item := range data {
		num, ok := item.(float64)
		if !ok {
			mapped[i] = item
			continue
		}

		switch operation {
		case "double":
			mapped[i] = num * 2
		case "square":
			mapped[i] = num * num
		case "negate":
			mapped[i] = -num
		default:
			mapped[i] = num
		}
	}

	return map[string]interface{}{
		"result": mapped,
	}, nil
}

func reduceArray(args map[string]interface{}) (interface{}, error) {
	data, ok := args["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("data must be an array")
	}

	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	if len(data) == 0 {
		return map[string]interface{}{"result": 0}, nil
	}

	var result float64
	switch operation {
	case "sum":
		result = 0
		for _, item := range data {
			if num, ok := item.(float64); ok {
				result += num
			}
		}
	case "product":
		result = 1
		for _, item := range data {
			if num, ok := item.(float64); ok {
				result *= num
			}
		}
	case "max":
		result = data[0].(float64)
		for _, item := range data {
			if num, ok := item.(float64); ok && num > result {
				result = num
			}
		}
	case "min":
		result = data[0].(float64)
		for _, item := range data {
			if num, ok := item.(float64); ok && num < result {
				result = num
			}
		}
	case "avg":
		sum := 0.0
		count := 0
		for _, item := range data {
			if num, ok := item.(float64); ok {
				sum += num
				count++
			}
		}
		if count > 0 {
			result = sum / float64(count)
		}
	}

	return map[string]interface{}{
		"result": result,
	}, nil
}

func sortArray(args map[string]interface{}) (interface{}, error) {
	data, ok := args["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("data must be an array")
	}

	order, ok := args["order"].(string)
	if !ok {
		order = "asc"
	}

	// Convert to float64 slice for sorting
	numbers := make([]float64, 0, len(data))
	for _, item := range data {
		if num, ok := item.(float64); ok {
			numbers = append(numbers, num)
		}
	}

	if order == "asc" {
		sort.Float64s(numbers)
	} else {
		sort.Sort(sort.Reverse(sort.Float64Slice(numbers)))
	}

	// Convert back to interface slice
	sorted := make([]interface{}, len(numbers))
	for i, num := range numbers {
		sorted[i] = num
	}

	return map[string]interface{}{
		"result": sorted,
	}, nil
}

func groupBy(args map[string]interface{}) (interface{}, error) {
	data, ok := args["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("data must be an array")
	}

	key, ok := args["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key must be a string")
	}

	groups := make(map[string][]interface{})
	for _, item := range data {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		keyValue, ok := obj[key].(string)
		if !ok {
			keyValue = "unknown"
		}

		groups[keyValue] = append(groups[keyValue], item)
	}

	return map[string]interface{}{
		"result": groups,
	}, nil
}

func mergeArrays(args map[string]interface{}) (interface{}, error) {
	arrays, ok := args["arrays"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("arrays must be an array of arrays")
	}

	merged := []interface{}{}
	for _, arr := range arrays {
		if subArray, ok := arr.([]interface{}); ok {
			merged = append(merged, subArray...)
		}
	}

	return map[string]interface{}{
		"result": merged,
		"count":  len(merged),
	}, nil
}

func uniqueValues(args map[string]interface{}) (interface{}, error) {
	data, ok := args["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("data must be an array")
	}

	seen := make(map[interface{}]bool)
	unique := []interface{}{}

	for _, item := range data {
		// Use string representation for map key
		key := fmt.Sprintf("%v", item)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		}
	}

	return map[string]interface{}{
		"result": unique,
		"count":  len(unique),
	}, nil
}
