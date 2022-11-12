package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.szostok.io/version"
	"go.szostok.io/version/printer"
)

func NOTmain() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// get the user options
	var addr string
	var token string
	var v bool
	var h bool

	flag.StringVar(&addr, "addr", "localhost:3000", "ws host")
	flag.StringVar(&token, "token", "", "auth token")
	flag.BoolVar(&v, "version", false, "print version")
	flag.BoolVar(&v, "v", false, "print version")
	flag.BoolVar(&v, "help", false, "print options")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if v {
		var verPrinter = printer.New()
		var info = version.Get()
		if err := verPrinter.PrintInfo(os.Stdout, info); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/results/" + token}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("dial error: %s", err.Error())
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		var filenamePrefix = []byte("filepath:")
		var sumPrefix = []byte("sha512:")
		var shaDecoder = sha512.New()

		var filename string
		var imageBytes []byte
		var sha string
		for {
			msgType, message, err := c.ReadMessage()
			if err != nil {
				log.Errorf("read message error: %s", err.Error())
				return
			}

			// we expect three messages per image:
			// 1. the filename
			// 2. the image
			// 3. the sha512 of the image
			if msgType == websocket.TextMessage {
				if bytes.HasPrefix(message, filenamePrefix) {
					filename = string(bytes.TrimPrefix(message, filenamePrefix))
				} else if bytes.HasPrefix(message, sumPrefix) {
					sha = string(bytes.TrimPrefix(message, sumPrefix))
				}
			} else if msgType == websocket.BinaryMessage {
				imageBytes = message
			}

			// once we have all the messages, validate the hash and write the image
			if filename != "" && len(imageBytes) > 0 && sha != "" {
				shaDecoder.Write(imageBytes)
				var sum = hex.EncodeToString(shaDecoder.Sum(nil))

				if sum == sha {
					err = os.WriteFile(filename, message, os.ModePerm)
					if err != nil {
						log.Errorf("error writing image to disk, err: %s", err.Error())
					}
					log.Info("wrote file: ", filename)
				} else {
					log.Errorf("sha512 sums do not match, theirs: %s, ours: %s", sha, sum)
				}

				// reset for next one
				filename = ""
				sha = ""
				imageBytes = imageBytes[:0]
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Errorf("error sending heartbeat message, err: %s", err.Error())
				return
			}
		case <-interrupt:
			log.Info("shutting down")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Errorf("error sending close message, err: %s", err.Error())
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
