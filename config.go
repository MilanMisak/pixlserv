package main

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/ReshNesh/go-colorful"
	"gopkg.in/yaml.v1"
)

const (
	// LRU = Least recently used
	LRU = "LRU"
	// LFU = Least frequently used
	LFU = "LFU"
)

const (
	defaultThrottlingRate             = 60 // Requests per min
	defaultCacheLimit                 = 0  // No. of bytes
	defaultJpegQuality                = 75
	defaultUploadMaxFileSize          = 5 * 1024 * 1024 // No. of bytes
	defaultAllowCustomTransformations = true
	defaultAllowCustomScale           = true
	defaultAsyncUploads               = false
	defaultAuthorisedGet              = false
	defaultAuthorisedUpload           = false
	defaultLocalPath                  = "local-images"
	defaultCacheStrategy              = LRU
	defaultFont                       = "fonts/DejaVuSans.ttf"
)

var (
	// Config is a global configuration object
	Config Configuration
)

// Configuration specifies server configuration options
type Configuration struct {
	throttlingRate, cacheLimit, jpegQuality, uploadMaxFileSize                                  int
	allowCustomTransformations, allowCustomScale, asyncUploads, authorisedGet, authorisedUpload bool
	localPath, cacheStrategy                                                                    string
	transformations                                                                             map[string]Transformation
	eagerTransformations                                                                        []Transformation
}

func configInit(configFilePath string) error {
	Config = Configuration{defaultThrottlingRate, defaultCacheLimit, defaultJpegQuality, defaultUploadMaxFileSize, defaultAllowCustomTransformations, defaultAllowCustomScale, defaultAsyncUploads, defaultAuthorisedGet, defaultAuthorisedUpload, defaultLocalPath, defaultCacheStrategy, make(map[string]Transformation), make([]Transformation, 0)}

	if configFilePath == "" {
		return nil
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(data), &m)

	throttlingRate, ok := m["throttling-rate"].(int)
	if ok && throttlingRate >= 0 {
		Config.throttlingRate = throttlingRate
	}

	jpegQuality, ok := m["jpeg-quality"].(int)
	if ok && jpegQuality >= 1 && jpegQuality <= 100 {
		Config.jpegQuality = jpegQuality
	}

	uploadMaxFileSize, ok := m["upload-max-file-size"].(int)
	if ok && uploadMaxFileSize > 0 {
		Config.uploadMaxFileSize = uploadMaxFileSize
	}

	allowCustomTransformations, ok := m["allow-custom-transformations"].(bool)
	if ok {
		Config.allowCustomTransformations = allowCustomTransformations
	}

	allowCustomScale, ok := m["allow-custom-scale"].(bool)
	if ok {
		Config.allowCustomScale = allowCustomScale
	}

	asyncUploads, ok := m["async-uploads"].(bool)
	if ok {
		Config.asyncUploads = asyncUploads
	}

	authorisation, ok := m["authorisation"].(map[interface{}]interface{})
	if ok {
		get, ok := authorisation["get"].(bool)
		if ok {
			Config.authorisedGet = get
		}
		upload, ok := authorisation["upload"].(bool)
		if ok {
			Config.authorisedUpload = upload
		}
	}

	localPath, ok := m["local-path"].(string)
	if ok {
		Config.localPath = localPath
	}

	cache, ok := m["cache"].(map[interface{}]interface{})
	if ok {
		limit, ok := cache["limit"].(int)
		if ok {
			Config.cacheLimit = limit
		}

		strategy, ok := cache["strategy"].(string)
		if ok && (strategy == LRU || strategy == LFU) {
			Config.cacheStrategy = strategy
		}
	}

	transformations, ok := m["transformations"].([]interface{})
	if !ok {
		return nil
	}

	for _, transformationInterface := range transformations {
		transformation, ok := transformationInterface.(map[interface{}]interface{})
		if !ok {
			continue
		}

		parametersStr, ok := transformation["parameters"].(string)
		if !ok {
			continue
		}

		params, err := parseParameters(parametersStr)
		if err != nil {
			return fmt.Errorf("invalid transformation parameters: %s (%s)", parametersStr, err)
		}

		name, ok := transformation["name"].(string)
		if !ok {
			continue
		}
		if !isValidTransformationName(name) {
			return fmt.Errorf("invalid transformation name: %s", name)
		}

		t := Transformation{&params, nil, make([]*Text, 0)}

		watermarkMap, ok := transformation["watermark"].(map[interface{}]interface{})
		if ok {
			imagePath, ok := watermarkMap["source"].(string)
			if !ok {
				return fmt.Errorf("a watermark needs to have a source specified")
			}
			// x and y will default to 0 if not found in config
			x := watermarkMap["x-pos"].(int)
			y := watermarkMap["y-pos"].(int)
			t.watermark = &Watermark{imagePath, x, y}
		}

		texts, ok := transformation["text"].([]interface{})
		if ok {

			for _, textMap := range texts {
				text, ok := textMap.(map[interface{}]interface{})
				if !ok {
					continue
				}

				content, ok := text["content"].(string)

				// x and y will default to 0 if not found in config
				x := text["x-pos"].(int)
				y := text["y-pos"].(int)

				colorStr, ok := text["color"].(string)
				if !ok {
					return fmt.Errorf("text needs to have a color specified")
				}
				color, err := colorful.Hex(colorStr)
				if err != nil {
					return err
				}

				font, ok := text["font"].(string)
				if !ok {
					font = defaultFont
				}
				//TODO - check that the file exists

				size, ok := text["size"].(int)
				if !ok {
					return fmt.Errorf("%v is not a valid size", text["size"])
				}
				if size < 1 {
					return fmt.Errorf("size needs to be at least 1")
				}

				t.texts = append(t.texts, &Text{content, font, x, y, size, color})
			}
		}

		Config.transformations[name] = t

		eager, ok := transformation["eager"].(bool)
		if ok && eager {
			Config.eagerTransformations = append(Config.eagerTransformations, t)
		}
	}

	return nil
}

var (
	transformationNameConfigRe = regexp.MustCompile("^([0-9A-Za-z-]+)$")
)

func isValidTransformationName(name string) bool {
	return transformationNameConfigRe.MatchString(name)
}
