package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	PARAMETER_WIDTH    = "w"
	PARAMETER_HEIGHT   = "h"
	PARAMETER_CROPPING = "c"
	PARAMETER_GRAVITY  = "g"
)

// TODO - create a params struct

// Turns a string like "w_400,h_300" into a map[w:400 h:300]
// The second return value is an error message
// Also validates the parameters to make sure they have valid values
// w = width, h = height
func parseParameters(parametersStr string) (map[string]string, error) {
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
				return nil, fmt.Errorf("Could not parse value for parameter: %q", key)
			}
			if value <= 0 {
				return nil, fmt.Errorf("Value %q must be > 0: %q", key, key)
			}
		case PARAMETER_CROPPING:
			value = strings.ToLower(value)
			if len(value) > 1 {
				return nil, fmt.Errorf("Value %q must have only 1 character", key)
			}
			if !isValidCroppingMode(value) {
				return nil, fmt.Errorf("Invalid value for %q", key)
			}
		case PARAMETER_GRAVITY:
			value = strings.ToLower(value)
			if len(value) > 2 {
				return nil, fmt.Errorf("Value %q must have at most 2 characters", key)
			}
			if !isValidGravity(value) {
				return nil, fmt.Errorf("Invalid value for %q", key)
			}
		}

		parameters[key] = value
	}

	_, croppingModePresent := parameters[PARAMETER_CROPPING]
	if !croppingModePresent {
		parameters[PARAMETER_CROPPING] = DEFAULT_CROPPING_MODE
	}

	_, gravityPresent := parameters[PARAMETER_GRAVITY]
	if !gravityPresent {
		parameters[PARAMETER_GRAVITY] = DEFAULT_GRAVITY
	}

	return parameters, nil
}

// Turns an image file path and a map of parameters into a file path combining both.
// It can then be used for file lookups.
// The function assumes that imagePath contains an extension at the end.
func createFilePath(imagePath string, parameters map[string]string) (string, error) {
	i := strings.LastIndex(imagePath, ".")
	if i == -1 {
		return "", fmt.Errorf("Invalid image path")
	}

	orderedKeys := make([]string, len(parameters))
	j := 0
	for k, _ := range parameters {
		orderedKeys[j] = k
		j++
	}
	sort.Strings(orderedKeys)

	paramsString := ""
	for _, v := range orderedKeys {
		if _, present := parameters[v]; present {
			if paramsString != "" {
				paramsString += ","
			}
			paramsString += v + "_" + parameters[v]
		}
	}

	return imagePath[:i] + "[" + paramsString + "]" + imagePath[i:], nil
}
