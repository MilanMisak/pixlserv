package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/throttled"
	"github.com/PuerkitoBio/throttled/store"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
)

type UploadForm struct {
	PhotoUpload *multipart.FileHeader `form:"photo"`
}

func main() {
	// Set up logging for CLI
	log.SetPrefix("")
	log.SetFlags(0) // Remove the timestamp

	// Connect to redis
	err := redisInit()
	if err != nil {
		log.Println("Connecting to redis failed", err)
	}

	app := cli.NewApp()
	app.Name = "pixlserv"
	app.Usage = "transform and serve images"
	app.Version = "1.0"
	app.Commands = []cli.Command{
		{
			Name:  "run",
			Usage: "Runs the server (run [config-file])",
			Action: func(c *cli.Context) {
				// Set up logging for server
				log.SetPrefix("[pixlserv] ")

				if len(c.Args()) < 1 {
					log.Println("You need to provide a path to a config file")
					return
				}
				configFilePath := c.Args().First()

				// Initialise configuration
				err := configInit(configFilePath)
				if err != nil {
					log.Println("Configuration reading failed:", err)
					return
				}
				log.Printf("Running with config: %+v", Config)

				// Initialise authentication
				err = authInit()
				if err != nil {
					log.Println("Authentication initialisation failed:", err)
					return
				}

				// Initialise storage
				err = storageInit()
				if err != nil {
					log.Println("Storage initialisation failed:", err)
					return
				}

				// Run the server
				m := martini.Classic()
				if Config.throttlingRate > 0 {
					m.Use(throttler(Config.throttlingRate))
				}
				m.Get("/((?P<apikey>[A-Z0-9]+)/)?image/:parameters/**", transformationHandler)
				m.Post("/((?P<apikey>[A-Z0-9]+)/)?upload", binding.MultipartForm(UploadForm{}), uploadHandler)
				go m.Run()

				// Wait for when the program is terminated
				ch := make(chan os.Signal)
				signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
				<-ch

				// Clean up
				redisCleanUp()
				storageCleanUp()
			},
		},
		{
			Name:  "api-key",
			Usage: "Manages API keys",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "Adds a new one",
					Action: func(c *cli.Context) {
						key, err := generateKey()
						if err != nil {
							log.Println("Adding a new API key failed, please try again")
							return
						}

						log.Println("Key added:", key)
					},
				},
				{
					Name:  "info",
					Usage: "Shows information about a key (info [key])",
					Action: func(c *cli.Context) {
						if len(c.Args()) < 1 {
							log.Println("You need to provide an existing key")
							return
						}
						key := c.Args().First()
						permissions, err := infoAboutKey(key)
						if err != nil {
							log.Println(err.Error())
							return
						}
						log.Println("Key:", key)
						log.Println("Permissions:", permissions)
					},
				},
				{
					Name:  "list",
					Usage: "Shows all keys",
					Action: func(c *cli.Context) {
						keys, err := listKeys()
						if err != nil {
							log.Println("Retrieving the list of all keys failed")
							return
						}

						log.Println("Keys:", keys)
					},
				},
				{
					Name:  "modify",
					Usage: "Modifies permissions for a key (modify [key] [add/remove] [" + authPermissionsOptions() + "])",
					Action: func(c *cli.Context) {
						if len(c.Args()) < 3 {
							log.Println("You need to provide an existing key, operation and a permission")
							return
						}
						key := c.Args().First()
						err := modifyKey(key, c.Args()[1], c.Args()[2])
						if err != nil {
							log.Println(err.Error())
							return
						}
						log.Println("The key has been updated")
					},
				},
				{
					Name:  "remove",
					Usage: "Removes an existing key (remove [key])",
					Action: func(c *cli.Context) {
						if len(c.Args()) < 1 {
							log.Println("You need to provide an existing key")
							return
						}
						err := removeKey(c.Args().First())
						if err != nil {
							log.Println(err.Error())
							return
						}
						log.Println("The key was successfully removed")
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

func transformationHandler(params martini.Params) (int, string) {
	if !hasPermission(params["apikey"], GetPermission) {
		return http.StatusUnauthorized, ""
	}

	var transformation Transformation
	transformationName := parseTransformationName(params["parameters"])
	if transformationName != "" {
		var ok bool
		transformation, ok = Config.transformations[transformationName]
		if !ok {
			return http.StatusBadRequest, "Unknown transformation: " + transformationName
		}
	} else if Config.allowCustomTransformations {
		parameters, err := parseParameters(params["parameters"])
		if err != nil {
			return http.StatusBadRequest, err.Error()
		}
		transformation = Transformation{&parameters, nil}
	} else {
		return http.StatusBadRequest, "Custom transformations not allowed"
	}
	baseImagePath, scale := parseBasePathAndScale(params["_1"])
	if Config.allowCustomScale {
		parameters := transformation.params.WithScale(scale)
		transformation.params = &parameters
	}

	// Check if the image with the given parameters already exists
	// and return it
	fullImagePath, _ := createFilePath(baseImagePath, transformation.params)
	img, format, err := loadFromCache(fullImagePath)
	if err == nil {
		var buffer bytes.Buffer
		writeImage(img, format, &buffer)

		return http.StatusOK, buffer.String()
	}

	// Load the original image and process it
	if !imageExists(baseImagePath) {
		return http.StatusNotFound, "Image not found: " + baseImagePath
	}

	img, format, err = loadImage(baseImagePath)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}

	imgNew := transformCropAndResize(img, &transformation)

	var buffer bytes.Buffer
	err = writeImage(imgNew, format, &buffer)
	if err != nil {
		log.Println("Writing an image to the response failed:", err)
	}

	// Cache the image asynchronously to speed up the response
	go func() {
		err = addToCache(fullImagePath, imgNew, format)
		if err != nil {
			log.Println("Saving an image to cache failed:", err)
		}
	}()

	return http.StatusOK, buffer.String()
}

func uploadHandler(params martini.Params, uf UploadForm) (int, string) {
	if !hasPermission(params["apikey"], UploadPermission) {
		return http.StatusUnauthorized, ""
	}

	file, err := uf.PhotoUpload.Open()
	if err != nil {
		return http.StatusBadRequest, err.Error()
	}

	limit := io.LimitReader(file, int64(Config.uploadMaxFileSize+1))
	data, err := ioutil.ReadAll(limit)
	if err != nil {
		return http.StatusBadRequest, err.Error()
	}
	if len(data) > Config.uploadMaxFileSize {
		return http.StatusBadRequest, "max file size exceeded"
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return http.StatusBadRequest, err.Error()
	}

	defer file.Close()

	// Not a big fan of .jpeg file extensions
	baseImagePath := fmt.Sprintf("%d.%s", time.Now().Unix(), strings.Replace(format, "jpeg", "jpg", 1))
	log.Printf("Uploading %s", baseImagePath)

	if Config.asyncUploads {
		go func() {
			saveImage(img, format, baseImagePath)
		}()
	} else {
		_, err := saveImage(img, format, baseImagePath)
		if err != nil {
			return http.StatusInternalServerError, err.Error()
		}
	}

	// Eager transformations
	if len(Config.eagerTransformations) > 0 {
		go func() {
			for _, transformation := range Config.eagerTransformations {
				imgNew := transformCropAndResize(img, &transformation)
				fullImagePath, _ := createFilePath(baseImagePath, transformation.params)
				addToCache(fullImagePath, imgNew, format)
			}
		}()
	}

	return http.StatusOK, ""
}

func throttler(perMinRate int) http.Handler {
	t := throttled.RateLimit(throttled.PerMin(perMinRate), &throttled.VaryBy{RemoteAddr: true}, store.NewMemStore(1000))
	return t.Throttle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Nothing needed here
	}))
}
