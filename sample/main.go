package main

import (
	"goil"
)

func main() {
	app := goil.Default()
	app.GET("/hello/:who", func(c *goil.Context) {
		who := c.Param("who")
		c.Text("hello," + who)
	})
	xrouter := app.XRouter()
	xrouter.GET("/greet/:who", func(p *struct {
		Who string `path:"who" validator:"reg(/^[a-zA-Z]{3,6}$/)"`
	}) string {
		return "hello," + p.Who
	})
	app.Run(":8081")
}
