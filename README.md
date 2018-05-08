# goil

```
    __________       
   / ________/         ______
  / / _____  _______  /__/  /  
 / / /____ \/  ___  \/  /  /   
/ /______/ /  /__/  /  /  /__ 
\_________/\_______/\_/\____/ 
```

my graduation project , a micro web framwork by golang


# example
```

type Account struct {
	Username string
	Password string
}

func main() {
	app := goil.Default()
	g := &goil.GroupX{
		ErrorHandler:  goil.DefErrHandler,
		RenderHandler: goil.DefRenderHandler,
	}

	app.POST("/login", g.Wrapper(func(a *Account) *Account {
		return a
	}))
	app.Run(":8081")
}

```



```go
package main

import (
	"goil"
	"goil/session"
	"mime/multipart"
	"os"
)

type Params struct {
	P1       string         `form:"p1" validator:"reg(^[a-z]*$)"`
	Int      int            `form:"int" validator:"range(0 150)"`
	PInt     *int           `form:"int" validator:"max(150) min(0)"`
	PP1      *string        `form:"p1"`
	PPP1     **string       `form:"p1"`
	Path     string         `path:"path"`
	PPPath   ***string      `path:"path"`
	Music    *****string    `file:"music"` //filename
	Size     ***int64       `file:"music"` //fileSize
	MemFile  multipart.File `file:"music"` //<=32M, cached in memory
	DiskFile *os.File       `file:"music"` //>32M, temp in disk
	File     *goil.File     `file:"music"`
}

func main() {
	session.EnableRdsSession()
	app := goil.Default()

	app.POST("/echo_params/:path", func(c *goil.Context) {
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

	if err := app.Run(":8081"); err != nil {
		panic(err)
	}

}
```