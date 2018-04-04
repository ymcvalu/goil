package main

import (
	"goil"
)

type Params struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	app := goil.New()
	app.GET("/", func(c *goil.Context) {
		c.String("hello,goil!")
	})
	app.GET("/json", func(c *goil.Context) {
		c.JSON(map[string]string{
			"name": "Jim",
			"age":  "19",
		})
	})
	app.GET("/indentJson", func(c *goil.Context) {
		c.IndentJSON(map[string]string{
			"name": "Jim",
			"age":  "19",
		})
	})
	app.GET("/securyJson", func(c *goil.Context) {
		c.SecuryJSON(map[string]string{
			"name": "Jim",
			"age":  "19",
		})
	})
	app.POST("/json/echo", func(c *goil.Context) {
		var params = Params{}
		c.Bind(&params)
		c.JSON(params)
	})
	app.Run(":8081")

}
