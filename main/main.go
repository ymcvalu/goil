package main

import (
	"goil"
)

func main() {
	app := goil.New()
	app.GET("/", func(c *goil.Context) {
		c.String("hello,goil!")
	})
	app.Run(":8081")
}
