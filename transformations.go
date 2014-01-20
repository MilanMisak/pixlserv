package main

const (
	CROPPING_MODE_EXACT     = "e"
	CROPPING_MODE_ALL       = "a"
	CROPPING_MODE_PART      = "p"
	CROPPING_MODE_KEEPSCALE = "k"

	DEFAULT_CROPPING_MODE = CROPPING_MODE_EXACT
)

func isValidCroppingMode(str string) bool {
	return str == CROPPING_MODE_EXACT || str == CROPPING_MODE_ALL || str == CROPPING_MODE_PART || str == CROPPING_MODE_KEEPSCALE
}
