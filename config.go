package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v1"
)

const (
	LRU = "LRU"
	LFU = "LFU"
)

const (
	defaultThrottlingRate             = 60              // Requests per min
	defaultCacheLimit                 = 0               // No. of bytes
	defaultUploadMaxFileSize          = 5 * 1024 * 1024 // No. of bytes
	defaultAllowCustomTransformations = true
	defaultAllowCustomScale           = true
	defaultAsyncUploads               = false
	defaultLocalPath                  = "local-images"
	defaultCacheStrategy              = LRU
)

var (
	config Config
)

type Config struct {
	throttlingRate, cacheLimit, uploadMaxFileSize              int
	allowCustomTransformations, allowCustomScale, asyncUploads bool
	localPath, cacheStrategy                                   string
	transformations                                            map[string]Params
	eagerTransformations                                       []Params
}

func configInit(configFilePath string) (Config, error) {
	config = Config{defaultThrottlingRate, defaultCacheLimit, defaultUploadMaxFileSize, defaultAllowCustomTransformations, defaultAllowCustomScale, defaultAsyncUploads, defaultLocalPath, defaultCacheStrategy, make(map[string]Params), make([]Params, 0)}

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

	uploadMaxFileSize, ok := m["upload-max-file-size"].(int)
	if ok && uploadMaxFileSize > 0 {
		config.uploadMaxFileSize = uploadMaxFileSize
	}

	allowCustomTransformations, ok := m["allow-custom-transformations"].(bool)
	if ok {
		config.allowCustomTransformations = allowCustomTransformations
	}

	allowCustomScale, ok := m["allow-custom-scale"].(bool)
	if ok {
		config.allowCustomScale = allowCustomScale
	}

	asyncUploads, ok := m["async-uploads"].(bool)
	if ok {
		config.asyncUploads = asyncUploads
	}

	localPath, ok := m["local-path"].(string)
	if ok {
		config.localPath = localPath
	}

	cache, ok := m["cache"].(map[interface{}]interface{})
	if ok {
		limit, ok := cache["limit"].(int)
		if ok {
			config.cacheLimit = limit
		}

		strategy, ok := cache["strategy"].(string)
		if ok && (strategy == LRU || strategy == LFU) {
			config.cacheStrategy = strategy
		}
	}

	transformations, ok := m["transformations"].([]interface{})
	if ok {
		for _, transformationInterface := range transformations {
			transformation, ok := transformationInterface.(map[interface{}]interface{})
			if ok {
				parametersStr, ok := transformation["parameters"].(string)
				if ok {
					params, err := parseParameters(parametersStr)
					if err != nil {
						return config, fmt.Errorf("invalid transformation parameters: %s (%s)", parametersStr, err)
					}
					name, ok := transformation["name"].(string)
					if ok {
						config.transformations[name] = params
						eager, ok := transformation["eager"].(bool)

						if ok && eager {
							config.eagerTransformations = append(config.eagerTransformations, params)
						}
					}
				}
			}
		}
	}

	return config, nil
}

func getConfig() Config {
	return config
}
