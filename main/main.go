package main

import (
	"goil"
	"os"
)

type Params struct {
	Name     *string `form:"name" validator:"reg(^[a-z]*$)"`
	Age      *int    `form:"age" validator:"range(0 150)"`
	FileName *string `file:"music"`
	Size     int64   `file:"music"`
	File     os.File `file:"music"`
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
		err := c.Bind(&params)

		if err != nil {
			c.JSON(map[string]string{
				"Msg": err.Error(),
			})
			return
		}
		c.JSON(params)

	})
	app.GET("/param/:name/:age", func(c *goil.Context) {
		if val, exist := c.Param("age"); exist {
			c.String(val)
		}
	})
	app.Run(":8081")
}
