package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	log "github.com/sirupsen/logrus"
	_ "golang.org/x/image/webp"
)

var base64Prefix = []byte("data:image/png;base64,")

func setupWebServer(originalImages, upsizedImages chan string, imageDir, username, password string) *fiber.App {
	app := fiber.New()

	app.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			username: password,
		},
	}))

	app.Use(logger.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Static("/upsized", "./upsized")

	app.Get("/about", func(c *fiber.Ctx) error {
		return c.SendString("about")
	})

	var shaDecoder = sha512.New()
	app.Post("/upsize", func(c *fiber.Ctx) error {
		var imagePath string
		var imageSHA string

		var form, err = c.MultipartForm()
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("body needs to be multipart form")
		}

		if imageShaArr := form.Value["sha512"]; len(imageShaArr) == 1 {
			imageSHA = imageShaArr[0]
		} else {
			return c.Status(http.StatusBadRequest).SendString("sha512 needs to be one value, was: " + strconv.Itoa(len(imageShaArr)))
		}

		if imageArr := form.File["image"]; len(imageArr) == 1 {
			// write the file
			log.Info(imageArr[0].Filename, imageArr[0].Size, imageArr[0].Header["Content-Type"][0])
			if err := c.SaveFile(imageArr[0], filepath.Join(imageDir, imageArr[0].Filename)); err != nil {
				log.Errorf("unable to write image to disk, err: %s", err.Error())
				return c.Status(http.StatusInternalServerError).SendString("unable to write image to disk, err: " + err.Error())
			}

			imagePath = filepath.Join(imageDir, imageArr[0].Filename)

			// open and decode image to make sure its actaully an image
			var fileHandle, err = os.Open(imagePath)
			defer fileHandle.Close()
			if err != nil {
				log.Errorf("unable to open image, err: %s", err.Error())
				return c.Status(http.StatusInternalServerError).SendString("unable to open image, err: " + err.Error())
			}
			_, _, err = image.Decode(fileHandle)
			if err != nil {
				log.Errorf("unable to decode image, err: %s", err.Error())
				return c.Status(http.StatusInternalServerError).SendString("unable to decode image, err: " + err.Error())
			}

			// verify the sha
			fileHandle.Seek(0, 0)
			imageBytes, err := ioutil.ReadAll(fileHandle)
			if err != nil {
				log.Errorf("unable to decode image, err: %s", err.Error())
				return c.Status(http.StatusInternalServerError).SendString("unable to decode image, err: " + err.Error())
			}
			shaDecoder.Write(imageBytes)
			var sum = hex.EncodeToString(shaDecoder.Sum(nil))
			if imageSHA != sum {
				return c.Status(http.StatusBadRequest).SendString(fmt.Sprintf("sha512 sums do not match, yours: %s, ours: %s", imageSHA, sum))
			}

		} else {
			return c.Status(http.StatusBadRequest).SendString("image needs to be one value, was: " + strconv.Itoa(len(imageArr)))
		}

		originalImages <- filepath.Join(imageDir, imagePath)
		return c.Status(http.StatusOK).SendString("queued")
	})

	app.Use("/results", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		var allowed, ok = c.Locals("allowed").(bool)
		if ok && allowed {

		}

		for upsizedImage := range upsizedImages {
			var imageBytes, err = ioutil.ReadFile(upsizedImage)
			if err != nil {
				log.Errorf("ws error: %s", err.Error())
			}
			var encodedImage []byte
			base64.StdEncoding.Encode(encodedImage, imageBytes)
			if err = c.WriteMessage(websocket.BinaryMessage, append(base64Prefix, encodedImage...)); err != nil {
				log.Errorf("ws error: %s", err.Error())
			}
		}

	}))

	return app
}
