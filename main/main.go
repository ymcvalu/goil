package main

import (
	"goil"
)

func main() {

	app := goil.Default()
	xr := app.XRouter()
	xr.POST("/login", func(p *Params) *Params {
		return p
	})
	app.Run(":8081")
}

type Params struct {
	Username string
	Password string
}
