package main

import (
	"goil"
)

func main() {

	type Params struct {
		Who string `path:"who" validator:"reg(/^[a-zA-Z]{3,6}$/)"`
	}

	app := goil.Default()
	app.GET("/hello/:who", func(c *goil.Context) {
		who := c.Param("who")
		c.Text("hello," + who)
	})

	xrouter := app.XRouter()

	xrouter.GET("/greet/:who", func(p *Params) string {
		return "hello," + p.Who
	})
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
