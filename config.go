package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v1"
)

const (
	defaultThrottlingRate = 60 // Requests per min
)

type Config struct {
	throttlingRate int
}

func configInit(configFilePath string) (Config, error) {
	config := Config{defaultThrottlingRate}

	if configFilePath == "" {
		return config, nil
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return config, err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(data), &m)

	throttlingRate, ok := m["throttling-rate"].(int)
	if ok && throttlingRate >= 0 {
		config.throttlingRate = throttlingRate
	}

	return config, nil
}
