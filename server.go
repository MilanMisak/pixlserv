package main

import "fmt"
import "os"
import "github.com/codegangsta/martini"

const (
	LOCAL_IMAGES_PATH = "local-images"
)

func main() {
	m := martini.Classic()
	m.Get("/image/**", func(params martini.Params) string {
		imagePath := params["_1"]

		log("Looking for: " + LOCAL_IMAGES_PATH + "/" + imagePath)

		if _, err := os.Stat(LOCAL_IMAGES_PATH + "/" + imagePath); os.IsNotExist(err) {
			return "image does not exist: " + imagePath
		} else {
			return "image exists: " + imagePath
		}
	})
	m.Run()
}

func log(message string) {
	fmt.Println("[pixlserv] " + message)
}
