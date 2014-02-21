package main

import (
    "testing"
)

func TestParseParameters(t *testing.T) {
    act, _ := parseParameters("w_400,h_300")
    exp := Params{400, 300, DEFAULT_CROPPING_MODE, DEFAULT_GRAVITY}
    if act != exp {
        t.Errorf("Expected: %v, actual: %v", exp, act)
    }

    act, _ = parseParameters("w_200,h_300,c_k,g_c")
    exp = Params{200, 300, CROPPING_MODE_KEEPSCALE, GRAVITY_CENTER}
    if act != exp {
        t.Errorf("Expected: %v, actual: %v", exp, act)
    }
}
