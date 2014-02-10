package main

import (
    "reflect"
    "testing"
)

func TestParseParameters(t *testing.T) {
    act, _ := parseParameters("w_400,h_300")
    exp := map[string]string {
        PARAMETER_WIDTH: "400",
        PARAMETER_HEIGHT: "300",
        PARAMETER_CROPPING: DEFAULT_CROPPING_MODE,
        PARAMETER_GRAVITY: DEFAULT_GRAVITY,
    }
    if !reflect.DeepEqual(exp, act) {
        t.Errorf("Expected: %v, actual: %v", exp, act)
    }

    act, _ = parseParameters("w_200,h_300,c_k,g_c")
    exp = map[string]string {
        PARAMETER_WIDTH: "200",
        PARAMETER_HEIGHT: "300",
        PARAMETER_CROPPING: CROPPING_MODE_KEEPSCALE,
        PARAMETER_GRAVITY: GRAVITY_CENTER,
    }
    if !reflect.DeepEqual(exp, act) {
        t.Errorf("Expected: %v, actual: %v", exp, act)
    }
}
