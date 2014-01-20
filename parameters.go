package main

import (
	"strconv"
	"strings"
)

const (
	PARAMETER_WIDTH    = "w"
	PARAMETER_HEIGHT   = "h"
	PARAMETER_CROPPING = "c"
)

// Turns a string like "w_400,h_300" into a map[w:400 h:300]
// The second return value is an error message
// Also validates the parameters to make sure they have valid values
// w = width, h = height
func parseParameters(parametersStr string) (map[string]string, string) {
	parameters := make(map[string]string)
	parts := strings.Split(parametersStr, ",")
	for _, part := range parts {
		keyAndValue := strings.SplitN(part, "_", 2)
		key := keyAndValue[0]
		value := keyAndValue[1]

		switch key {
		case PARAMETER_WIDTH, PARAMETER_HEIGHT:
			value, err := strconv.Atoi(value)
			if err != nil {
				return nil, "Could not parse value for parameter: " + key
			}
			if value <= 0 {
				return nil, "Value [" + key + "] must be > 0: " + key
			}
		case PARAMETER_CROPPING:
			value = strings.ToLower(value)
			if len(value) > 1 {
				return nil, "Value [" + key + "] must have only 1 character"
			}
			if !isValidCroppingMode(value) {
				value = DEFAULT_CROPPING_MODE
			}
		}

		parameters[key] = value
	}
	return parameters, ""
}
