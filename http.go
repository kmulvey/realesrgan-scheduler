package main

import (
	"crypto/sha512"
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
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	_ "golang.org/x/image/webp"
)

var base64Prefix = []byte("data:image/png;base64,")

func setupWebServer(originalImages, upsizedImages chan string, imageDir, username, password string) *fiber.App {
	app := fiber.New()

	app.Use("/upsize*", basicauth.New(basicauth.Config{
		Users: map[string]string{
			username: password,
		},
	}))
	app.Use(func(c *fiber.Ctx) error {
		if _, ok := c.Locals("username").(string); ok {
			var authToken = uuid.NewString()
			c.Locals("token", authToken)
			c.Set("token", authToken)
		}
		return c.Next()
	})

	app.Use(logger.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Static("/upsized", "./upsized")

	app.Post("/upsize", func(c *fiber.Ctx) error {
		var shaDecoder = sha512.New()
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

		originalImages <- imagePath
		return c.Status(http.StatusOK).SendString("queued")
	})

	app.Use("/results/:token", func(c *fiber.Ctx) error {
		if token, ok := c.Locals("token").(string); ok {
			if c.Params("token") != token {
				return c.SendStatus(http.StatusUnauthorized)
			}
		}
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/results/:token", websocket.New(func(c *websocket.Conn) {
		var shaDecoder = sha512.New()
		var allowed, ok = c.Locals("allowed").(bool)
		if ok && allowed {
			for upsizedImage := range upsizedImages {
				var imageBytes, err = ioutil.ReadFile(upsizedImage)
				if err != nil {
					log.Errorf("ws error: %s", err.Error())
					// TODO
				}

				// send the file name
				if err = c.WriteMessage(websocket.TextMessage, []byte("filepath:"+filepath.Base(upsizedImage))); err != nil {
					log.Errorf("ws error: %s", err.Error())
				}

				// send the image
				if err = c.WriteMessage(websocket.BinaryMessage, imageBytes); err != nil {
					log.Errorf("ws error: %s", err.Error())
				}

				// send the sha
				shaDecoder.Write(imageBytes)
				var sum = hex.EncodeToString(shaDecoder.Sum(nil))
				if err = c.WriteMessage(websocket.TextMessage, []byte("sha512:"+sum)); err != nil {
					log.Errorf("ws error: %s", err.Error())
				}
				shaDecoder.Reset()
			}
		}
	}))

	return app
}
