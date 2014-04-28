package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	parameterWidth    = "w"
	parameterHeight   = "h"
	parameterCropping = "c"
	parameterGravity  = "g"
)

// Params is a struct of parameters specifying an image transformation
type Params struct {
	width, height     int
	cropping, gravity string
}

// ToString turns parameters into a unique string for each possible assignment of parameters
func (p Params) ToString() string {
	return fmt.Sprintf("%s_%s,%s_%s,%s_%d,%s_%d", parameterCropping, p.cropping, parameterGravity, p.gravity, parameterHeight, p.height, parameterWidth, p.width)
}

// Turns a string like "w_400,h_300" into a Params struct
// The second return value is an error message
// Also validates the parameters to make sure they have valid values
// w = width, h = height
func parseParameters(parametersStr string) (Params, error) {
	params := Params{0, 0, DefaultCroppingMode, DefaultGravity}
	parts := strings.Split(parametersStr, ",")
	for _, part := range parts {
		keyAndValue := strings.SplitN(part, "_", 2)
		key := keyAndValue[0]
		value := keyAndValue[1]

		switch key {
		case parameterWidth, parameterHeight:
			value, err := strconv.Atoi(value)
			if err != nil {
				return params, fmt.Errorf("could not parse value for parameter: %q", key)
			}
			if value <= 0 {
				return params, fmt.Errorf("value %q must be > 0: %q", key, key)
			}
			if key == parameterWidth {
				params.width = value
			} else {
				params.height = value
			}
		case parameterCropping:
			value = strings.ToLower(value)
			if len(value) > 1 {
				return params, fmt.Errorf("value %q must have only 1 character", key)
			}
			if !isValidCroppingMode(value) {
				return params, fmt.Errorf("invalid value for %q", key)
			}
			params.cropping = value
		case parameterGravity:
			value = strings.ToLower(value)
			if len(value) > 2 {
				return params, fmt.Errorf("value %q must have at most 2 characters", key)
			}
			if !isValidGravity(value) {
				return params, fmt.Errorf("invalid value for %q", key)
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
		return "", fmt.Errorf("invalid image path")
	}

	return imagePath[:i] + "--" + parameters.ToString() + "--" + imagePath[i:], nil
}
