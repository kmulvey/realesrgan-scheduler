package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {

	var client = fiber.AcquireClient()
	var agent = client.Post("http://localhost:3000/upsize/")

	agent.BasicAuth("kmulvey", "nLXLGYvSrsnM29eH3ykAJxykHGxJsT")

	var req = agent.Request()
	req.Header.SetMethod(fiber.MethodPost)
	if err := agent.Parse(); err != nil {
		panic(err)
	}
	var args = fiber.AcquireArgs()
	args.Set("image_name", "test.jpg")
	agent.SendFile("image", "./test.jpg").MultipartForm(args)

	var resp = fiber.AcquireResponse()
	fmt.Println(agent.Do(agent.Request(), resp))
	fmt.Println(resp.StatusCode())
}
