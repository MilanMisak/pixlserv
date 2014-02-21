package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	PARAMETER_WIDTH    = "w"
	PARAMETER_HEIGHT   = "h"
	PARAMETER_CROPPING = "c"
	PARAMETER_GRAVITY  = "g"
)

type Params struct {
    width, height int
    cropping, gravity string
}

func (p Params) ToString() string {
    return fmt.Sprintf("%s_%s,%s_%s,%s_%d,%s_%d", PARAMETER_CROPPING, p.cropping, PARAMETER_GRAVITY, p.gravity, PARAMETER_HEIGHT, p.height, PARAMETER_WIDTH, p.width)
}

// Turns a string like "w_400,h_300" into a Params struct
// The second return value is an error message
// Also validates the parameters to make sure they have valid values
// w = width, h = height
func parseParameters(parametersStr string) (Params, error) {
    params := Params{0, 0, DEFAULT_CROPPING_MODE, DEFAULT_GRAVITY}
	parts := strings.Split(parametersStr, ",")
	for _, part := range parts {
		keyAndValue := strings.SplitN(part, "_", 2)
		key := keyAndValue[0]
		value := keyAndValue[1]

		switch key {
		case PARAMETER_WIDTH, PARAMETER_HEIGHT:
			value, err := strconv.Atoi(value)
			if err != nil {
				return params, fmt.Errorf("Could not parse value for parameter: %q", key)
			}
			if value <= 0 {
				return params, fmt.Errorf("Value %q must be > 0: %q", key, key)
			}
            if key == PARAMETER_WIDTH {
                params.width = value
            } else {
                params.height = value
            }
		case PARAMETER_CROPPING:
			value = strings.ToLower(value)
			if len(value) > 1 {
				return params, fmt.Errorf("Value %q must have only 1 character", key)
			}
			if !isValidCroppingMode(value) {
				return params, fmt.Errorf("Invalid value for %q", key)
			}
            params.cropping = value
		case PARAMETER_GRAVITY:
			value = strings.ToLower(value)
			if len(value) > 2 {
				return params, fmt.Errorf("Value %q must have at most 2 characters", key)
			}
			if !isValidGravity(value) {
				return params, fmt.Errorf("Invalid value for %q", key)
			}
            params.gravity = value
		}
	}

	return params, nil
}

// Turns an image file path and a map of parameters into a file path combining both.
// It can then be used for file lookups.
// The function assumes that imagePath contains an extension at the end.
func createFilePath(imagePath string, parameters Params) (string, error) {
	i := strings.LastIndex(imagePath, ".")
	if i == -1 {
		return "", fmt.Errorf("Invalid image path")
	}

	return imagePath[:i] + "[" + parameters.ToString() + "]" + imagePath[i:], nil
}
