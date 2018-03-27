package main

import (
	"fmt"
	"goil"
)

func main() {
	app := goil.New()
	app.GET("/", func(c *goil.Context) {
		fmt.Println("invoke")
	})
	app.Run(":8081")
}
